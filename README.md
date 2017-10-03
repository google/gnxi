
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](https://godoc.org/github.com/google/gnxi?status.svg)](https://godoc.org/github.com/google/gnxi)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gnxi)](https://goreportcard.com/report/github.com/google/gnxi)
[![Build Status](https://travis-ci.org/google/gnxi.svg?branch=master)](https://travis-ci.org/google/gnxi)
[![codecov.io](https://codecov.io/github/google/gnxi/coverage.svg?branch=master)](https://codecov.io/github/google/gnxi?branch=master)

# gNxI Tools

gNxi - gRPC Network Management/Operations Interface

A collection of tools for Network Management that use the gNMI and gNOI protocols.

### Summary

_Note_: These tools are intended for testing and as reference implementation of the protocol.

*  [gNMI Capabilities](./gnmi_capabilities)
*  [gNMI Set](./gnmi_get)
*  [gNMI Get](./gnmi_set)
*  [gNMI Target](./gnmi_target)

### Documentation

*  See [gNMI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnmi).
*  See [gNOI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnoi).
*  See [Openconfig documentation](http://www.openconfig.net/).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See Docker for instructions on how to test against network equipment.

### Prerequisites

Install __go__ in your system https://golang.org/doc/install.

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

## Docker

[FAUCET](https://github.com/faucetsdn/faucet) currently includes a [Dockerfile](https://github.com/faucetsdn/faucet/blob/master/Dockerfile.gnmi) to setup the environment that facilitates testing these tools against network equipment. See the [gNMI FAUCET documentation](https://github.com/faucetsdn/faucet/tree/master/gnmi) for more information.

## Disclaimer

*  This is not an official Google product.
*  See [how to contribute](CONTRIBUTING.md).
