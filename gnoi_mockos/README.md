# gNOI Mock OS Generator

A simple shell binary generates a Mock OS package to use with the gNOI_target.

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
  -unsupported false
```
