#!/bin/bash

##################################
### Demo of oc_config_validate ###
##################################

GNMI_PORT=9339
NO_TLS=0
ROOT_CA=0
CLIENT_TLS=0
STOP_ON_ERROR=0
LOG_GNMI=0
DEBUG=0

BASEDIR=$(dirname $0)

# start_gnmi_target <gnmi_port>
start_gnmi_target() {

    # From https://github.com/google/gnxi/tree/master/gnmi_target

    if [[ ! -x "$(go env GOPATH)/bin/gnmi_target" ]]; then
        go install github.com/google/gnxi/gnmi_target@latest
    fi

    OPTS="-key $BASEDIR/certs/target.key -cert $BASEDIR/certs/target.crt -ca $BASEDIR/certs/ca.crt"
    if [[ "$NO_TLS" -eq 1 ]]; then
        OPTS="--notls"
    fi

    echo "--- Start TARGET $OPTS"
    $(go env GOPATH)/bin/gnmi_target -bind_address ":$1" -config $BASEDIR/target_config.json --insecure $OPTS >> /dev/null 2>&1 &
    sleep 3
}

# stop_gnmi_target <gnmi_port>
stop_gnmi_target() {
    echo "--- Stop TARGET"
    pkill -f "$(go env GOPATH)/bin/gnmi_target -bind_address :$1"
}

# start_oc_config_validate <gnmi_port>
start_oc_config_validate() {
    OPTS=""
    if [[ "$NO_TLS" -eq 1 ]]; then
        OPTS="--no_tls"
    fi
    if [[ "$ROOT_CA" -eq 1 ]]; then
        OPTS="-ca $BASEDIR/certs/ca.crt"
    fi
    if [[ "$CLIENT_TLS" -eq 1 ]]; then
        OPTS="$OPTS -key $BASEDIR/certs/client.key -cert $BASEDIR/certs/client.crt"
    fi
    if [[ "$STOP_ON_ERROR" -eq 1 ]]; then
        OPTS="$OPTS --stop_on_error"
    fi
    if [[ "$LOG_GNMI" -eq 1 ]]; then
        OPTS="$OPTS --log_gnmi"
    fi
    if [[ "$DEBUG" -eq 1 ]]; then
        OPTS="$OPTS --debug"
    fi
    echo "--- Start oc_config_validate $OPTS"
    PYTHONPATH=$PYTHONPATH:$BASEDIR/.. python3 -m oc_config_validate --target "localhost:$1" --tests_file $BASEDIR/tests.yaml --results_file $BASEDIR/results.json --init_config_file $BASEDIR/init_config.json --init_config_xpath "/system/config" $OPTS

  }

parse_options() {
  while getopts "p:NRCSLDh" opt; do
    case ${opt} in
      p )
        GNMI_PORT=$OPTARG
        ;;
      N )
        NO_TLS=1
        ;;
      R )
        ROOT_CA=1
        ;;
      C )
        CLIENT_TLS=1
        ;;
      S )
        STOP_ON_ERROR=1
        ;;
      L )
        LOG_GNMI=1
        ;;
      D )
        DEBUG=1
        ;;
      * ) 
        echo "
demo.sh [-p <gNMI Port>]
        [-N] # No TLS
        [-R] # Use Root CA file
        [-C] # Use client TLS files
        [-S] # Stop on error
        [-L] # Log Gnmi messages to the test results
        [-D] # Enable debug output
"
          return 1
        ;;
    esac
  done
return 0
}

main() {
    if parse_options "$@"; then
        if [[ ! ( -f $BASEDIR/certs/target.key && -f $BASEDIR/certs/target.crt  ) ]]; then
          echo "--- Creating local self-signed certificates"
          mkdir -p $BASEDIR/certs
          cd $BASEDIR/certs
          wget https://raw.githubusercontent.com/google/gnxi/master/certs/generate.sh -O - -o /dev/null | bash
        fi
        start_gnmi_target "${GNMI_PORT}"
        start_oc_config_validate "${GNMI_PORT}"
        stop_gnmi_target "${GNMI_PORT}"
    fi
}

main "$@"
