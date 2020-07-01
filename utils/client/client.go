package client

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/google/gnxi/utils/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GnoiEncrypted creates an encrypted unauthenticated connection to the target.
func GnoiEncrypted(c tls.Certificate, targetAddr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{c},
			RootCAs:            nil,
		}))}

	conn, err := grpc.Dial(targetAddr, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GnoiAuthenticated creates an encrypted authenticated connection to the target.
func GnoiAuthenticated(caEnt *entity.Entity, targetAddr, targetName string) (*grpc.ClientConn, error) {
	clientEnt, err := entity.CreateSigned("client", nil, caEnt)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(caEnt.Certificate.Leaf)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			ServerName:   targetName,
			Certificates: []tls.Certificate{*clientEnt.Certificate},
			RootCAs:      caPool,
		}))}

	conn, err := grpc.Dial(targetAddr, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
