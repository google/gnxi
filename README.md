
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

gNMI Clients:

*  [gNMI Capabilities](./gnmi_capabilities)
*  [gNMI Get](./gnmi_get)
*  [gNMI Set](./gnmi_set)
*  gNMI Subscribe - in progress

gNMI Targets:

*  [gNMI Target](./gnmi_target)

gNOI Clients

*  [gNOI Cert](./gnoi_cert)
*  [gNOI OS](./gnoi_os)
*  [gNOI Reset](./gnoi_reset)

gNOI Targets

*  [gNOI Target](./gnoi_target)

Helpers

*  [gNOI mockOS](./gnoi_mockos)
*  [certificate generator](./certs)

### Documentation

*  See [gNMI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnmi).
*  See [gNOI Protocol documentation](https://github.com/openconfig/gnoi).
*  See [Openconfig documentation](http://www.openconfig.net/).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See Docker for instructions on how to test against network equipment.

### Prerequisites

Install __go__ in your system https://golang.org/doc/install. Requires golang1.7+.

### Clone

Clone the project to your __go__ source folder:
```
mkdir -p $GOPATH/src/github.com/google/
cd $GOPATH/src/github.com/google/
git clone https://github.com/google/gnxi.git
```

### Running

To run the binaries:

```
cd $GOPATH/src/github.com/google/gnxi/gnmi_get
go run ./gnmi_get.go
```

## Disclaimer

*  This is not an official Google product.
*  See [how to contribute](CONTRIBUTING.md).
