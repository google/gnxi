package main

import (
	"crypto/x509"
	"fmt"

	"github.com/google/go-cmp/cmp"
)

func main() {
	// cmpOpts = []cmp.Option{cmpopts.IgnoreUnexported(sync.RWMutex{}), cmp.AllowUnexported(Manager{}), cmpopts.IgnoreUnexported(CertInfo{}), cmpopts.IgnoreUnexported(x509.Certificate{})} //, cmpopts.IgnoreTypes(&x509.Certificate{})}
	cmpOpts := []cmp.Option{}

	caBundle1 := []x509.Certificate{x509.Certificate{Version: 1}, x509.Certificate{Version: 2}}
	caBundle2 := []x509.Certificate{x509.Certificate{}}

	fmt.Printf("Bu: (-want +got):\n%s", cmp.Diff(caBundle1, caBundle2, cmpOpts...))

}
