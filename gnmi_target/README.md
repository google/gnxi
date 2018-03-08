# gNMI Target

A simple shell binary that implements a gNMI Target with in-memory configuration and telemetry.

## Install

```
go get github.com/google/gnxi/gnmi_target
go install github.com/google/gnxi/gnmi_target
```

## Run

```
gnmi_target \
  -bind_address :10161 \
  -config openconfig-openflow.json \
  -key server.key \
  -cert server.crt \
  -ca ca.crt \
  -username foo \
  -password bar \
  -alsologtostderr
```
