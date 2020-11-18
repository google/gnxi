/* Copyright 2018 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Binary implements a Certificate Management service client.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/gnxi/gnoi/cert"
	credUtils "github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/entity"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/golang/glog"
)

var (
	certID     = flag.String("cert_id", "", "Certificate Management certificate ID.")
	certIDs    = flag.String("cert_ids", "", "Comma separated list of Certificate Management certificate IDs for revoke operation")
	op         = flag.String("op", "get", "Certificate Management operation, one of: provision, install, rotate, get, revoke, check")
	targetCN   = credUtils.TargetName
	targetAddr = flag.String("target_addr", "localhost:9339", "The target address in the format of host:port")
	timeOut    = flag.Duration("time_out", 5*time.Second, "Timeout for the operation, 5 seconds by default")
	minKeySize = flag.Uint("min_key_size", 1024, "Minimum key size")
	country    = flag.String("country", "CH", "Country in CSR parameters")
	state      = flag.String("state", "ZRH", "State in CSR parameters")
	org        = flag.String("organization", "OpenConfig", "Organization in CSR parameters")
	orgUnit    = flag.String("organizational_unit", "gNxI", "Organizational unit in CSR parameters")
	ipAddress  = flag.String("ip_address", "127.0.0.1", "IP address in CSR parameters")

	caEnt     *entity.Entity
	ctx       context.Context
	cancel    func()
	dial      = grpc.Dial
	loadCerts = credUtils.LoadCertificates
)

func main() {
	flag.Parse()

	if *targetCN == "" {
		log.Exit("Must set a Common Name ID with -target_name.")
	}

	ctx, cancel = context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	switch *op {
	case "provision":
		caEnt = credUtils.GetCAEntity()
		certIDCheck()
		provision()
	case "install":
		caEnt = credUtils.GetCAEntity()
		certIDCheck()
		install()
	case "rotate":
		caEnt = credUtils.GetCAEntity()
		certIDCheck()
		rotate()
	case "revoke":
		revoke()
	case "check":
		check()
	case "get":
		get()
	default:
		log.Exitf("Unknown operation: %q", *op)
	}
}

func certIDCheck() {
	if *certID == "" {
		log.Exit("Must set a certificate ID with -cert_id.")
	}
}

// gnoiEncrypted creates an encrypted TLS connection to the target.
func gnoiEncrypted(c tls.Certificate) (*grpc.ClientConn, *cert.Client) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{c},
			RootCAs:            nil,
		}))}

	conn, err := dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Failed dial to %q: %v", *targetAddr, err)
	}

	client := cert.NewClient(conn)
	return conn, client
}

// gnoiAuthenticated creates an authenticated TLS connection to the target.
func gnoiAuthenticated(targetName string) (*grpc.ClientConn, *cert.Client) {
	signed, certPool := loadCerts()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			ServerName:   targetName,
			Certificates: signed,
			RootCAs:      certPool,
		}))}

	conn, err := dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Failed dial to %q: %v", *targetAddr, err)
	}

	client := cert.NewClient(conn)
	return conn, client
}

// signer is called to create a Certificate from a CSR.
func signer(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	e, err := entity.FromSigningRequest(csr)
	if err != nil {
		return nil, fmt.Errorf("failed generating a cert from a CSR: %v", err)
	}
	if err := e.SignWith(caEnt); err != nil {
		return nil, fmt.Errorf("failed to sign the certificate: %v", err)
	}
	return e.Certificate.Leaf, nil
}

// provision provisions a target in bootstrapping mode.
func provision() {
	// Using the CA x509 cert as default Certificate, but can be any.
	conn, client := gnoiEncrypted(*caEnt.Certificate)
	defer conn.Close()
	pkiName := pkix.Name{CommonName: *targetCN, Organization: []string{*org}, OrganizationalUnit: []string{*orgUnit}, Country: []string{*country}, Province: []string{*state}}

	if err := client.Install(ctx, *certID, uint32(*minKeySize), pkiName, *ipAddress, signer, []*x509.Certificate{caEnt.Certificate.Leaf}); err != nil {
		log.Exit("Failed Install:", err)
	}
	log.Info("Install success")
}

// install installs a certificate in authenticated mode.
func install() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()
	pkiName := pkix.Name{CommonName: *targetCN, Organization: []string{*org}, OrganizationalUnit: []string{*orgUnit}, Country: []string{*country}, Province: []string{*state}}

	if err := client.Install(ctx, *certID, uint32(*minKeySize), pkiName, *ipAddress, signer, []*x509.Certificate{caEnt.Certificate.Leaf}); err != nil {
		log.Exit("Failed Install:", err)
	}
	log.Info("Install success")
}

// rotate rotates a certificate in authenticated mode.
func rotate() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()
	pkiName := pkix.Name{CommonName: *targetCN, Organization: []string{*org}, OrganizationalUnit: []string{*orgUnit}, Country: []string{*country}, Province: []string{*state}}

	if err := client.Rotate(ctx, *certID, uint32(*minKeySize), pkiName, *ipAddress, signer, []*x509.Certificate{caEnt.Certificate.Leaf}, func() error { return nil }); err != nil {
		log.Exit("Failed Rotate:", err)
	}
	log.Info("Rotate success")
}

// revoke revokes a certificate in authenticated mode.
func revoke() {
	var revokeCertIDs = []string{*certID}

	if *certIDs != "" {
		revokeCertIDs = strings.FieldsFunc(*certIDs, func(r rune) bool { return r == ',' })
		if len(revokeCertIDs) == 0 {
			log.Exit("Must specify comma separated certificate IDs when using -cert_ids")
		}
	} else if *certID == "" {
		log.Exit("Must set a certificate ID with -cert_id or set multiple IDs with -cert_ids")
	}
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	revoked, err := client.RevokeCertificates(ctx, revokeCertIDs)
	if err != nil {
		log.Exit("Failed RevokeCertificates:", err)
	}
	log.Info("RevokeCertificates:\n", pretty.Sprint(revoked))
}

// revoke checks if a target can generate certificates - authenticated mode.
func check() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	resp, err := client.CanGenerateCSR(ctx)
	if err != nil {
		log.Exit("Failed CanGenerateCSR:", err)
	}
	log.Info("CanGenerateCSR:\n", pretty.Sprint(resp))
}

// get fetches the installed certificates on a target - authenticated mode.
func get() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	resp, err := client.GetCertificates(ctx)
	if err != nil {
		log.Exit("Failed GetCertificates:", err)
	}

	pretty.DefaultFormatter[reflect.TypeOf(&x509.Certificate{})] = func(c *x509.Certificate) string {
		return pretty.Sprint(c.Subject.CommonName)
	}
	log.Info("GetCertificates:\n", pretty.Sprint(resp))
}
