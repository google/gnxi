# gNMI Set

A simple shell binary that performs a SET against a gNMI Target.

## Install

```
go get github.com/google/gnxi/gnmi_set
go install github.com/google/gnxi/gnmi_set
```

## Run

Run gnmi\_set -help to see usage. For example:

```
gnmi_set \
  -delete /system/openflow/agent/config/max-backoff \
  -replace /system/clock:@clock-config.json \
  -replace /system/openflow/agent/config/max-backoff:12 \
  -update /system/clock/config/timezone-name:"US/New York" \
  -target_addr localhost:9339 \
  -target_name target.com \
  -key client.key \
  -cert client.crt \
  -ca ca.crt \
  -alsologtostderr
```
