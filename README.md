
![GitHub](https://img.shields.io/github/license/google/gnxi?style=for-the-badge)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=for-the-badge)](https://godoc.org/github.com/google/gnxi)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gnxi?style=for-the-badge)](https://goreportcard.com/report/github.com/google/gnxi)
![Build Status](https://img.shields.io/travis/google/gnxi?style=for-the-badge)
![Code coverage master](https://img.shields.io/codecov/c/github/google/gnxi/master?style=for-the-badge)

# gNxI Tools

*   gNMI - gRPC Network Management Interface
*   gNOI - gRPC Network Operations Interface

A collection of tools for Network Management that use the gNMI and gNOI protocols.

### Summary

_Note_: These tools are intended for testing and as reference implementation of the protocol.

#### gNMI Clients:

*  [gNMI Capabilities](./gnmi_capabilities)
*  [gNMI Get](./gnmi_get)
*  [gNMI Set](./gnmi_set)
*  gNMI Subscribe - in progress

#### gNMI Targets:

*  [gNMI Target](./gnmi_target)

#### gNOI Clients

*  [gNOI Cert](./gnoi_cert)
*  [gNOI OS](./gnoi_os)
*  [gNOI Reset](./gnoi_reset)

#### gNOI Targets

*  [gNOI Target](./gnoi_target)

#### Helpers

*  [gNOI mockOS](./gnoi_mockos)
*  [certificate generator](./certs)

### Documentation

*  See [gNMI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnmi).
*  See [gNOI Protocol documentation](https://github.com/openconfig/gnoi).
*  See [Openconfig documentation](http://www.openconfig.net/).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine.

### Prerequisites

Install __go__ in your system https://golang.org/doc/install. Requires golang1.7+.

### Download sources

```
go get -v github.com/google/gnxi/...
```

### Building and installing binaries

```
cd $GOPATH
mkdir bin
go install github.com/google/gnxi/...
ls -la $GOPATH\bin
```

### Generating certificates

```
cd $GOPATH\bin
./../src/github.com/google/gnxi/certs/generate.sh
```

### Running a client

```
cd $GOPATH\bin
./gnoi_reset \
-target_addr localhost:9399 \
-target_name target.com \
-rollback_os \
-zero_fill \
-key client.key \
-cert client.crt \
-ca ca.crt
```

Optionally define $GOBIN as $GOPATH\bin and add it to your path to run the binaries from any folder.

```
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

## Disclaimer

*  This is not an official Google product.
*  See [how to contribute](CONTRIBUTING.md).
