# gNMI Target

A simple shell binary that behaves as a gNMI Target.

Supported features:
*  Reflect a gNMI Get request.

## Install

```
go get github.com/samribeiro/gnmi/gnmi_target
go install github.com/samribeiro/gnmi/gnmi_target
```

## Run

```
gnmi_target \
  -bind_address :32123 \
  -key server.key \
  -cert server.crt \
  -ca ca.crt \
  -alsologtostderr
```
