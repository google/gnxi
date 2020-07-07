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

// Package gnoi contains required services for running a gnoi server.
package gnoi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"

	"github.com/google/gnxi/gnoi/cert"
	"github.com/google/gnxi/gnoi/os"
	"github.com/google/gnxi/gnoi/reset"
	"github.com/google/gnxi/utils/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rsaBitSize = 2048
)

// Server represents a target.
type Server struct {
	certServer         *cert.Server
	certManager        *cert.Manager
	defaultCertificate *tls.Certificate
	resetServer        *reset.Server
	osServer           *os.Server
}

// NewServer returns a new server that can be used by the mock target.
func NewServer(privateKey crypto.PrivateKey, defaultCertificate *tls.Certificate, resetSettings *reset.Settings, notifyReset reset.Notifier, osSettings *os.Settings) (*Server, error) {
	if defaultCertificate == nil {
		if privateKey == nil {
			var err error
			privateKey, err = rsa.GenerateKey(rand.Reader, rsaBitSize)
			if err != nil {
				return nil, fmt.Errorf("failed to generate private key: %v", err)
			}
		}
		e, err := entity.CreateSelfSigned("gNOI server", privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create self signed certificate: %v", err)
		}
		defaultCertificate = e.Certificate
	}

	certManager := cert.NewManager(defaultCertificate.PrivateKey)
	certServer := cert.NewServer(certManager)
	resetServer := reset.NewServer(resetSettings, notifyReset)
	osServer := os.NewServer(osSettings)
	return &Server{
		certServer:         certServer,
		certManager:        certManager,
		defaultCertificate: defaultCertificate,
		resetServer:        resetServer,
		osServer:           osServer,
	}, nil
}

// PrepareEncrypted prepares a gRPC server with the CertificateManagement service
// running with encryption but without authentication.
func (s *Server) PrepareEncrypted() *grpc.Server {

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAnyClientCert,
		Certificates: []tls.Certificate{*s.defaultCertificate},
		ClientCAs:    nil,
	}))}
	return grpc.NewServer(opts...)
}

// PrepareAuthenticated prepares a gRPC server with the CertificateManagement service
// running with full encryption and authentication.
func (s *Server) PrepareAuthenticated() *grpc.Server {
	config := func(*tls.ClientHelloInfo) (*tls.Config, error) {
		tlsCerts, x509Pool := s.certManager.TLSCertificates()
		return &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: tlsCerts,
			ClientCAs:    x509Pool,
		}, nil
	}
	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{GetConfigForClient: config}))}
	return grpc.NewServer(opts...)
}

// Register all implemented gRPC services.
func (s *Server) Register(g *grpc.Server) {
	s.certServer.Register(g)
	s.resetServer.Register(g)
	s.osServer.Register(g)
}

// RegCertificateManagement registers only the Certificate Management service in the gRPC Server.
func (s *Server) RegCertificateManagement(g *grpc.Server) {
	s.certServer.Register(g)
}

// RegisterCertNotifier registers a function that will be called everytime the number
// of Certificates or CA Certificates changes.
func (s *Server) RegisterCertNotifier(f cert.Notifier) {
	s.certManager.RegisterNotifier(f)
}
