# gNOI Target

A simple shell binary that implements a gNOI Target with the Certificate
Management service, supporting bootstrapping mode.

## Certificate Management service

This service provides a set of RPCs to Install, Rotate & Revoke Certificates and
CA Bundles in a Target. See [gNOI Cert proto definition](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto) for more.

## Bootstrapping mode

This target starts in bootstrapping mode allowing any encrypted TLS connection
to install certificates and CA bundles. For creating this encrypted connection
this target automatically creates a private key and a default self signed
Certificate.

Once a Certificate and a CA Certificate bundle is installed via the gNOI service
the Target changes its connection to authenticated mode. In this mode, only
authenticated TLS connections using the gNOI installed Certificates and CA
bundle, are allowed.



## Certificates and Key types supported

This Target currently only supports x509 Certificates and RSA Keys.

## Install

```
go get github.com/google/gnxi/gnoi_target
go install github.com/google/gnxi/gnoi_target
```

## Run

```
gnoi_target \
  -bind_address :10161 \
  -alsologtostderr
```
