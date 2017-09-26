# gNMI Get

A simple shell binary that performs a GET against a gNMI Target.

## Install

```
go get github.com/samribeiro/gnmi/gnmi_get
go install github.com/samribeiro/gnmi/gnmi_get
```

## Run

```
gnmi_get \
  -target_address localhost:32123 \
  -key client.key \
  -cert client.crt \
  -ca ca.crt \
  -target_name server \
  -alsologtostderr \
  -query "system/openflow/controllers/controller[name=main]/connections/connection[aux-id=0]/state/address"
```
