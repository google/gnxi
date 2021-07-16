# gNMI CLI

A simple Python script that performs interactions with gNMI Targets.

## Dependencies

The dependencies can be installed to either (1) your host's python environment or (2) within a python virtual environment.

System Python Installation
```
pip install -r requirements.txt

# Setup example
pip install setuptools
pip install --user --no-binary=protobuf -I grpcio-tools==1.15.0 grpcio==1.18.0
```
You may also need to pip install setuptools. You do not need "futures==3.2.0" if using Python3.

Virtualenv Installation
```
# install virtualenv
pip install virtualenv

# create a virtual environment
virtualenv venv
. venv/bin/activate
pip install -r requirements.txt
```

## Usage Examples
gNMI GetRequests. Substitute where applicable.
```
python py_gnmicli.py -m get -t <gnmi_target_addr> -p <port> -x <xpath> -user <user> -pass <password> [Optional] -rcert <CA certificate>
```
gNMI GetRequest for an Access Point target, who's hostname is "ap-1", with an xpath of root for access-points model. This gNMI Target is using a self-signed certificate, therefore we pass the --get_cert (-g) option, accompanied by the --host_override (-o). This automatically fetches the Certificate from the Target for building the gRPC channel. Host override is due to the Target self-signed certificate not including the IP/Hostname in the CN or SAN.
```
python py_gnmicli.py -m get -t example.net -x /access-points/access-point[hostname=ap-1]/ -user admin -pass admin -p 8080 -o openconfig.example.com
```
gNMI GetRequest for an Access Point target, who's hostname is "ap-1", with an xpath of the config container of Radio with ID 0. This Target is using a valid signed certificate from a public CA; therefore no need for -o or -g options:
```
python py_gnmicli.py -m get -t example.net -x /access-points/access-point[hostname=ap-1]/radios/radio[id=0]/config -user admin -pass admin -p 8080
```
gNMI SetRequest Replace for an Access Point target, who's hostname is "ap-1", with an xpath of the channel config leaf of Radio with ID 0 (This would assign channel 165 to this Radio). This Target is using a certificate which was signed by an internal CA; therefore we are providing the -rcert option, which provides the internal CA cert to the client for use in building the secure gRPC channel. This internal CA cert was obtained off-line:
```
python py_gnmicli.py -t example.net -p 443 -m set-replace -x /access-points/access-point[hostname=test-ap1]/radios/radio[id=0]/config/channel -user admin -pass admin -rcert ca.cert.pem -val 165
```
The above SetRequest Replace would output the following to stdout:
```
Performing SetRequest Replace, encoding=JSON_IETF  to  openconfig.example.com with the following gNMI Path
 -------------------------
 elem {
  name: "access-points"
}
elem {
  name: "access-point"
  key {
    key: "hostname"
    value: "test-ap1"
  }
}
elem {
  name: "radios"
}
elem {
  name: "radio"
  key {
    key: "id"
    value: "0"
  }
}
elem {
  name: "config"
}
elem {
  name: "channel"
}

The SetRequest response is below
-------------------------
 response {
  path {
    elem {
      name: "access-points"
    }
    elem {
      name: "access-point"
      key {
        key: "hostname"
        value: "test-ap1"
      }
    }
    elem {
      name: "radios"
    }
    elem {
      name: "radio"
      key {
        key: "id"
        value: "0"
      }
    }
    elem {
      name: "config"
    }
    elem {
      name: "channel"
    }
  }
  op: REPLACE
}
```
If the gNMI Target is utilizing a self-signed certificate it may also be required to supply the hostname utilized in the certificate (option: --host_override)

For example:
```
python py_gnmicli.py -t target1.example.com -p 443 -m get -x /access-points/access-point[hostname=test-ap1]/radios/radio[id=0]/config -o openconfig.mojonetworks.com -user admin -pass admin -g
```

### Notable Options
* The default output of a GetRequest is to dump the value as JSON. This can be changed with the -f flag.
* Pay special attention when utilizing a JSON file as the val when performing SetRequests. It MUST be preceded by an '@'; else it is assumed that you are providing a leaf value directly.
* The host_override (-o) option is most likely needed, if the Target is utilizing a self-signed certificate (unless the root CA is trusted on the host machine).
* Use the debug flag (-d) when troubleshooting/reporting gRPC errors.
* If the target isn't using TLS tunnels with its gRPC comms, use the notls (-n) option. This sends gRPC Metadata in clear-text, and should only be used for testing purposes.

## Docker
Docker image with Python3; based on version py_gnmicli version 0.4:
```
docker run --rm -it mike909/py_gnmicli:v0.4 python /gnxi/gnmi_cli_py/py_gnmicli.py -t <target_ip> -p <target_port> -x <xpath> -m get -user <user> -pass <password>
```
Or use the included Dockerfile to build your own.
