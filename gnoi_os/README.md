# gNOI OS Client

A simple shell binary that performs OS client operations against a gNOI target.

## gNOI OS Client Operations

* `-op install` installs the provided OS image onto the target.

* `-op activate` tells the target to boot into the specified OS version on next reboot.

* `-op verify` verifies the version of the OS currenly running on the target.

## Install

```
go get github.com/google/gnxi/gnoi_os
go install github.com/google/gnxi/gnoi_os
```

## Run 
```
gnoi_os \
-target_addr localhost:9399 \
-target_name target.com \
-ca ca.crt \
-key client.key \
-cert client.crt \
-version 1.1 \
-os myosfile.img \
-op install | activate | verify
```
