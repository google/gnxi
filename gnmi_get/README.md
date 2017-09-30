# gNMI Get

A simple shell binary that performs a GET against a gNMI Target.

## Install

```
go get github.com/google/gnxi/gnmi_get
go install github.com/google/gnxi/gnmi_get
```

## Run

```
gnmi_get \
  -target_addr localhost:10161 \
  -key client.key \
  -cert client.crt \
  -ca ca.crt \
  -target_name www.example.com \
  -alsologtostderr \
  -xpath "/system/openflow/controllers/controller[name=main]/connections/connection[aux-id=0]/state/address"
```
