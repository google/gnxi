# gNOI Mock OS Generator

A simple shell binary that generates a Mock OS package to use with the gNOI_target.

See [mockOS proto](./../utils/mockos/pb/mockos.proto) for more details.

## Install

```
go get github.com/google/gnxi/gnoi_mockos
go install github.com/google/gnxi/gnoi_mockos
```

## Run

```
gnoi_cert \
  -file ./myos.img \
  -version 1.10a \
  -size 100M \
  -activation_fail_message "I failed to activate" \
  -incompatible false \
```
