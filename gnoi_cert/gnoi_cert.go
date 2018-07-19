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
	"time"

	"github.com/google/gnxi/gnoi/cert"
	"github.com/google/gnxi/utils/entity"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/golang/glog"
)

var (
	certID     = flag.String("cert_id", "", "Certificate Management certificate ID.")
	op         = flag.String("op", "get", "Certificate Management operation, one of: provision, install, rotate, get, revoke, check")
	ca         = flag.String("ca", "", "CA certificate file.")
	key        = flag.String("key", "", "Private key file.")
	targetCN   = flag.String("target_name", "", "Common Name of the target.")
	targetAddr = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	timeOut    = flag.Duration("time_out", 5*time.Second, "Timeout for the operation, 5 seconds by default")

	caEnt  *entity.Entity
	ctx    context.Context
	cancel func()
)

func main() {
	flag.Parse()

	if *ca == "" || *key == "" {
		log.Exit("-ca and -key must be set with file locations")
	}
	if *targetCN == "" {
		log.Exit("Must set a Common Name ID with -targetCN.")
	}

	var err error
	if caEnt, err = entity.FromFile(*ca, *key); err != nil {
		log.Exitf("Failed to load certificate and key from file: %v", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	switch *op {
	case "provision":
		certIDCheck()
		provision()
	case "install":
		certIDCheck()
		install()
	case "rotate":
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

	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Failed dial to %q: %v", *targetAddr, err)
	}

	client := cert.NewClient(conn)
	return conn, client
}

// gnoiAuthenticated creates an authenticated TLS connection to the target.
func gnoiAuthenticated(targetName string) (*grpc.ClientConn, *cert.Client) {
	clientEnt, err := entity.CreateSigned("client", nil, caEnt)
	if err != nil {
		log.Exitf("Failed to create a signed entity: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(caEnt.Certificate.Leaf)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			ServerName:   targetName,
			Certificates: []tls.Certificate{*clientEnt.Certificate},
			RootCAs:      caPool,
		}))}

	conn, err := grpc.Dial(*targetAddr, opts...)
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

	if err := client.Install(ctx, *certID, pkix.Name{CommonName: *targetCN}, signer, []*x509.Certificate{caEnt.Certificate.Leaf}); err != nil {
		log.Exit("Failed Install:", err)
	}
	log.Info("Install success")
}

// install installs a certificate in authenticated mode.
func install() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	if err := client.Install(ctx, *certID, pkix.Name{CommonName: *targetCN}, signer, []*x509.Certificate{caEnt.Certificate.Leaf}); err != nil {
		log.Exit("Failed Install:", err)
	}
	log.Info("Install success")
}

// rotate rotates a certificate in authenticated mode.
func rotate() {
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	if err := client.Rotate(ctx, *certID, pkix.Name{CommonName: *targetCN}, signer, []*x509.Certificate{caEnt.Certificate.Leaf}, func() error { return nil }); err != nil {
		log.Exit("Failed Rotate:", err)
	}
	log.Info("Rotate success")
}

// revoke revokes a certificate in authenticated mode.
func revoke() {
	if *certID == "" {
		log.Exit("Must set a certificate ID with -cert_id.")
	}
	conn, client := gnoiAuthenticated(*targetCN)
	defer conn.Close()

	revoked, err := client.RevokeCertificates(ctx, []string{*certID})
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
