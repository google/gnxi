# gNOI Target

A shell binary that implements a gNOI Target supporting OS, Cert, Reset services
and [Simplified Bootstrapping](https://github.com/openconfig/gnoi/blob/master/docs/simplified_bootstrapping.md).

## Certificate Management service

This service provides a set of RPCs to Install, Rotate & Revoke Certificates and
CA Bundles in a Target. See [gNOI Cert proto definition](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto) for more.

## OS service

This service provides RPCs to Install, Activate and Verify OS installation on a Target.
See [gNOI OS proto definition](https://github.com/openconfig/gnoi/blob/master/os/os.proto) for more.

## Reset service

This service provides an RPC to Start a factory reset of the device. This includes
resetting all certificates on the device and setting it to bootstrapped mode.
See [gNOI Reset proto definition](https://github.com/openconfig/gnoi/blob/master/factory_reset/reset.proto) for more.

## Bootstrapping mode

If no target certificate and key are provided this target starts in bootstrapping
mode allowing any encrypted TLS connection to install certificates and CA bundles.
For creating this encrypted connection this target automatically creates a private
key and a default self signed Certificate.

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
./gnoi_target \
  -bind_address :9339 \
  -reset_unsupported true \
  -zero_fill_unsupported true \
  -factoryOS_version 1.0.0b \
  -installedVersions 1.0.1a 2.0.3b \
  -alsologtostderr
```
