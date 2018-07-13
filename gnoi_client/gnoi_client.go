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

	"github.com/google/gnxi/gnoi"
	"github.com/google/gnxi/utils/entity"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/golang/glog"
)

var (
	myCA       *entity.Entity
	myClientCA *entity.Entity
	myTargetCA *entity.Entity
	conString  = "127.0.0.1:45444"
	client     *gnoi.CertClient
)

func sign(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	e, err := entity.FromSigningRequest(csr)
	if err != nil {
		return nil, fmt.Errorf("failed FromSigningRequest: %v", err)
	}
	if err := e.SignWith(myTargetCA); err != nil {
		return nil, fmt.Errorf("failed SignWith: %v", err)
	}
	return e.Certificate.Leaf, nil
}

func install(id, cn string) {
	if err := client.Install(context.Background(), id, pkix.Name{CommonName: cn}, sign, []*x509.Certificate{myClientCA.Certificate.Leaf}); err != nil {
		log.Error("Failed Install:", err)
	} else {
		log.Info("Install success")
	}
}

func rotate(id, cn string) {
	if err := client.Rotate(context.Background(), id, pkix.Name{CommonName: cn}, sign, []*x509.Certificate{myClientCA.Certificate.Leaf}, func() error { return nil }); err != nil {
		log.Error("Failed Rotate:", err)
	} else {
		log.Info("Rotate success")
	}
}

func getCerts() {
	certs, err := client.GetCertificates(context.Background())
	if err != nil {
		log.Error("Failed GetCertificates:", err)
	} else {
		log.Info("GetCertificates:\n", pretty.Sprint(certs))
	}
}

func canGenSCR() {
	can, err := client.CanGenerateCSR(context.Background())
	if err != nil {
		log.Error("Failed CanGenerateCSR:", err)
	}
	log.Infof("CanGenerateCSR: %v", can)
}

func revoke(r []string) {
	revoked, err := client.RevokeCertificates(context.Background(), r)
	if err != nil {
		log.Error("Failed RevokeCertificates:", err)
	} else {
		log.Info("RevokeCertificates:\n", pretty.Sprint(revoked))
	}
}

func gnoiEncrypted(c tls.Certificate) *grpc.ClientConn {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "targetCert",
			Certificates:       []tls.Certificate{c},
			RootCAs:            nil,
		}))}

	conn, err := grpc.Dial(conString, opts...)
	if err != nil {
		log.Fatal("Failed dial to %q: %v", conString, err)
	}

	client = gnoi.NewCertClient(conn)
	return conn
}

func gnoiAuthenticated(c tls.Certificate) *grpc.ClientConn {
	caPool := x509.NewCertPool()
	caPool.AddCert(myTargetCA.Certificate.Leaf)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			ServerName:   "targetCert",
			Certificates: []tls.Certificate{c},
			RootCAs:      caPool,
		}))}

	conn, err := grpc.Dial(conString, opts...)
	if err != nil {
		log.Fatal("Failed dial to %q: %v", conString, err)
	}

	client = gnoi.NewCertClient(conn)
	return conn
}

func main() {
	flag.Parse()
	log.Info("Starting gNOI client.")

	pretty.DefaultFormatter[reflect.TypeOf(&x509.Certificate{})] = func(c *x509.Certificate) string {
		return pretty.Sprint(c.Subject.CommonName)
	}

	var err error
	myCA, err = entity.CreateSelfSigned("CA", nil)
	if err != nil {
		log.Fatal("Failed CA:", err)
	}
	myClientCA, err = entity.CreateSignedCA("ClientCA", nil, myCA)
	if err != nil {
		log.Fatal("Failed ClientCA:", err)
	}
	myTargetCA, err = entity.CreateSignedCA("TargetCA", nil, myCA)
	if err != nil {
		log.Fatal("Failed TargetCA:", err)
	}

	clientCert, err := entity.CreateSigned("clientCert", nil, myClientCA)
	if err != nil {
		log.Fatal("Failed clientCert:", err)
	}

	encryptedConn := gnoiEncrypted(*clientCert.Certificate)
	install("id1", "targetCert")
	encryptedConn.Close()

	time.Sleep(3 * time.Second)

	authenticatedConn := gnoiAuthenticated(*clientCert.Certificate)
	getCerts()
	rotate("id1", "targetCert")
	install("id2", "muCert")
	install("id3", "muCert")
	revoke([]string{"id2"})
	getCerts()

	authenticatedConn.Close()
	log.Info("Graceful exit.")
}
