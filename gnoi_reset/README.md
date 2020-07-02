# gNOI Factory Reset Client

A simple shell binary that performs Factory Reset client operations against a gNOI target.

## gNOI Factory Reset Options
*   `-rollback_os` will attempt to roll back the OS to the factory version and reset all certificates on the target. The target will then enter bootstrapping mode
*   `-zero_fill` will attempt to zero fill the device’s persistent storage.

## Install
```
go get github.com/google/gnxi/gnoi_reset
go install github.com/google/gnxi/gnoi_reset
```

## Run 
```
gnoi_reset \
-target_addr localhost:9399 \
-rollback_os \
-zero_fill \
-key client.key \
-cert client.crt \
-ca ca.crt
```