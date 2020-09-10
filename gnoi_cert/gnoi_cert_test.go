package main

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/google/gnxi/utils/entity"
	"google.golang.org/grpc"
)

func TestGnoiEncrypted(t *testing.T) {
	expectConn := &grpc.ClientConn{}
	dial = func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return expectConn, nil
	}
	conn, client := gnoiEncrypted(tls.Certificate{})
	if expectConn != conn {
		t.Errorf("Invalid connection (-want +got): -%v, +%v", expectConn, conn)
	}
	if client == nil {
		t.Error("Client was nil")
	}
}

func TestGnoiAuthenticated(t *testing.T) {
	expectConn := &grpc.ClientConn{}
	dial = func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return expectConn, nil
	}
	loadCerts = func() ([]tls.Certificate, *x509.CertPool) {
		return []tls.Certificate{}, &x509.CertPool{}
	}
	conn, client := gnoiAuthenticated("test")
	if expectConn != conn {
		t.Errorf("Invalid connection (-want +got): -%v, +%v", expectConn, conn)
	}
	if client == nil {
		t.Error("Client was nil")
	}
}

func TestSigner(t *testing.T) {
	caEnt = &entity.Entity{Template: &x509.Certificate{}}
	csr := &x509.CertificateRequest{}
	if leaf, err := signer(csr); err != nil || leaf == nil {
		t.Errorf("Leaf certificate was nil: %v", err)
	}
}