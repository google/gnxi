# gNMI Set

A simple shell binary that performs a SET against a gNMI Target.

## Install

```
go get github.com/google/gnxi/gnmi_set
go install github.com/google/gnxi/gnmi_set
```

## Run

```
gnmi_set \
  -config openconfig-openflow.json \
  -target_addr localhost:10161 \
  -target_name hostname.com \
  -key client.key \
  -cert client.crt \
  -ca ca.crt \
  -username foo
  -password bar
  -alsologtostderr
```
