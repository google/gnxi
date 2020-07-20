/* Copyright 2017 Google Inc.

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

// Package credentials loads certificates and validates user credentials.
package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/entity"
)

var (
	ca             = flag.String("ca", "", "CA certificate file.")
	caKey          = flag.String("ca_key", "", "CA private key file.")
	cert           = flag.String("cert", "", "Certificate file.")
	key            = flag.String("key", "", "Private key file.")
	insecure       = flag.Bool("insecure", false, "Skip TLS validation.")
	notls          = flag.Bool("notls", false, "Disable TLS validation. If true, no need to specify TLS related options.")
	authorizedUser = userCredentials{}
	usernameKey    = "username"
	passwordKey    = "password"
	caEnt          *entity.Entity
	caCert         *x509.Certificate
)

func init() {
	flag.StringVar(&authorizedUser.username, "username", "", "If specified, uses username/password credentials.")
	flag.StringVar(&authorizedUser.password, "password", "", "The password matching the provided username.")
}

type userCredentials struct {
	username string
	password string
}

func (a *userCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		usernameKey: a.username,
		passwordKey: a.password,
	}, nil
}

func (a *userCredentials) RequireTransportSecurity() bool {
	return true
}

// loadFromFile loads the certificates from files.
func loadFromFile() ([]tls.Certificate, *x509.CertPool) {
	certificate, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.Exitf("could not load client key pair: %s", err)
	}

	certPool := x509.NewCertPool()
	caFile, err := ioutil.ReadFile(*ca)
	if err != nil {
		log.Exitf("could not read CA certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(caFile); !ok {
		log.Exit("failed to append CA certificate")
	}

	block, _ := pem.Decode(caFile)
	caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Exit("Error parsing CA certificate", err)
	}
	return []tls.Certificate{certificate}, certPool
}

// generateFromCA generates a client certificate from the provided CA.
func generateFromCA() ([]tls.Certificate, *x509.CertPool) {
	GetCAEntity()
	clientEnt, err := entity.CreateSigned("client", nil, caEnt)
	if err != nil {
		log.Exitf("Failed to create a signed entity: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(caEnt.Certificate.Leaf)
	return []tls.Certificate{*clientEnt.Certificate}, caPool
}

// LoadCertificates loads certificates from files or generates them from the CA.
func LoadCertificates() ([]tls.Certificate, *x509.CertPool) {
	if *ca != "" {
		if *cert != "" && *key != "" {
			return loadFromFile()
		}
		if *caKey != "" {
			return generateFromCA()
		}
	}
	return []tls.Certificate{}, &x509.CertPool{}
}

// ClientCredentials generates gRPC DialOptions for existing credentials.
func ClientCredentials(server string) []grpc.DialOption {
	opts := []grpc.DialOption{}

	if *notls {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsConfig := &tls.Config{}
		if *insecure {
			tlsConfig.InsecureSkipVerify = true
		} else {
			certificates, certPool := LoadCertificates()
			if len(certificates) == 0 {
				log.Exit("Please provide -ca & -key or -ca, -cert & -ca_key")
			}
			tlsConfig.ServerName = server
			tlsConfig.Certificates = certificates
			tlsConfig.RootCAs = certPool
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	if authorizedUser.username != "" {
		return append(opts, grpc.WithPerRPCCredentials(&authorizedUser))
	}
	return opts
}

// GetCAEntity gets a CA entity from a CA file and private key.
func GetCAEntity() *entity.Entity {
	if caEnt != nil {
		return caEnt
	}
	if *caKey == "" {
		log.Exit("-ca_key must be set with file locations")
	}
	var err error
	if caEnt, err = entity.FromFile(*ca, *caKey); err != nil {
		log.Exitf("Failed to load certificate and key from file: %v", err)
	}
	caCert = caEnt.Certificate.Leaf
	return caEnt
}

// ServerCredentials generates gRPC ServerOptions for existing credentials.
func ServerCredentials() []grpc.ServerOption {
	if *notls {
		return []grpc.ServerOption{}
	}

	certificates, certPool := LoadCertificates()

	if *insecure {
		return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.VerifyClientCertIfGiven,
			Certificates: certificates,
			ClientCAs:    certPool,
		}))}
	}

	return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: certificates,
		ClientCAs:    certPool,
	}))}
}

// GetCACert returns the parsed CA cert.
func GetCACert() *x509.Certificate {
	return caCert
}

// AuthorizeUser checks for valid credentials in the context Metadata.
func AuthorizeUser(ctx context.Context) (string, bool) {
	authorize := false
	if authorizedUser.username == "" {
		authorize = true
	}
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "no Metadata found", authorize
	}
	user, ok := headers[usernameKey]
	if !ok || len(user) == 0 {
		return "no username in Metadata", authorize
	}
	pass, ok := headers[passwordKey]
	if !ok || len(pass) == 0 {
		return fmt.Sprintf("found username \"%s\" but no password in Metadata", user[0]), authorize
	}
	if authorize || pass[0] == authorizedUser.password && user[0] == authorizedUser.username {
		return fmt.Sprintf("authorized with \"%s:%s\"", user[0], pass[0]), true
	}
	return fmt.Sprintf("not authorized with \"%s:%s\"", user[0], pass[0]), false
}
