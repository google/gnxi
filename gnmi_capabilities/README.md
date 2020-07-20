# gNMI Capabilities

A simple shell binary that requests for Capabilities from a gNMI Target.

## Install

```
go get github.com/google/gnxi/gnmi_capabilities
go install github.com/google/gnxi/gnmi_capabilities
```

## Run

```
./gnmi_capabilities \
  -target_addr localhost:9339 \
  -target_name target.com \
  -key client.key \
  -cert client.crt \
  -ca ca.crt \
  -alsologtostderr
```
