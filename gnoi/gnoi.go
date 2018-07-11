// Package gnoi contains required services for running a gnoi server.
package gnoi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"

	"github.com/google/gnxi/utils/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	rsaBitSize = 2048
)

// Server does blah.
type Server struct {
	privateKey  crypto.PrivateKey
	certServer  *CertServer
	certManager *CertManager
}

// NewServer does blah.
func NewServer(privateKey crypto.PrivateKey) (*Server, error) {
	if privateKey == nil {
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, rsaBitSize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %v", err)
		}
	}
	certManager := NewCertManager(privateKey)
	certServer := NewCertServer(certManager)
	return &Server{
		privateKey:  privateKey,
		certServer:  certServer,
		certManager: certManager,
	}, nil
}

// PrepareEncrypted prepares a gRPC server with the CertificateManagement service
// running with encryption but without authentication.
func (s *Server) PrepareEncrypted() (*grpc.Server, error) {
	e, err := entity.CreateSelfSigned("gNOI server", s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create self signed certificate: %v", err)
	}
	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAnyClientCert,
		Certificates: []tls.Certificate{*e.Certificate},
		ClientCAs:    nil,
	}))}
	return grpc.NewServer(opts...), nil
}

// PrepareAuthenticated prepares a gRPC server with the CertificateManagement service
// running with full encryption and authentication.
func (s *Server) PrepareAuthenticated() *grpc.Server {
	config := func(*tls.ClientHelloInfo) (*tls.Config, error) {
		tlsCerts, x509Pool := s.certManager.Certificates()
		return &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: tlsCerts,
			ClientCAs:    x509Pool,
		}, nil
	}
	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{GetConfigForClient: config}))}
	return grpc.NewServer(opts...)
}

// RegCertificateManagement registers the Certificate Management service in the gRPC Server.
func (s *Server) RegCertificateManagement(g *grpc.Server) {
	s.certServer.Register(g)
}

// func Client(targetAddr string, certificates []tls.Certificate, caPool *x509.CertPool) error {
// 	opts := []grpc.DialOption{}
// 	tlsConfig := &tls.Config{}
//
// 	tlsConfig.InsecureSkipVerify = true
// 	tlsConfig.ServerName = "server"
// 	tlsConfig.Certificates = certificates
// 	tlsConfig.RootCAs = caPool
// 	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
//
// 	conn, err := grpc.Dial(targetAddr, opts...)
// 	if err != nil {
// 		log.Exitf("Dialing to %q failed: %v", targetAddr, err)
// 	}
// 	defer conn.Close()
//
// 	client := pb.NewCertificateManagementClient(conn)
//
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
//
// 	_, err = client.GetCertificates(ctx, &pb.GetCertificatesRequest{})
// 	if err != nil {
// 		log.Errorf("Hello failed: %v", err)
// 	} else {
// 		log.Info("Received Hello")
// 	}
//
// 	return err
// }
