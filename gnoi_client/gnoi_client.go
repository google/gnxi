package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"reflect"

	"github.com/google/gnxi/gnoi"
	"github.com/google/gnxi/utils/entity"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/golang/glog"
)

var (
	myCA      *entity.Entity
	conString = "127.0.0.1:45444"
)

func sign(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	e, err := entity.FromSigningRequest(csr)
	if err != nil {
		return nil, fmt.Errorf("failed FromSigningRequest: %v", err)
	}
	if err := e.SignWith(myCA); err != nil {
		return nil, fmt.Errorf("failed SignWith: %v", err)
	}
	return e.Certificate.Leaf, nil
}

func main() {
	flag.Parse()
	log.Info("Starting gNOI client.")

	pretty.DefaultFormatter[reflect.TypeOf(&x509.Certificate{})] = func(c *x509.Certificate) string {
		return pretty.Sprint(c.Subject)
	}

	var err error
	myCA, err = entity.CreateSelfSigned("MyCA", nil)
	if err != nil {
		log.Fatal("Failed CreateSelfSigned:", err)
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "server",
			Certificates:       []tls.Certificate{*myCA.Certificate},
			RootCAs:            nil,
		}))}

	conn, err := grpc.Dial(conString, opts...)
	if err != nil {
		log.Fatal("Failed dial to %q: %v", conString, err)
	}
	defer conn.Close()

	client := gnoi.NewCertClient(conn)

	// CanGenerateCSR
	can, err := client.CanGenerateCSR(context.Background())
	if err != nil {
		log.Error("Failed CanGenerateCSR:", err)
	}
	log.Infof("CanGenerateCSR: %v", can)

	// Install
	params := pkix.Name{CommonName: "First Certificate"}
	if err = client.Install(context.Background(), "hello certificate", params, sign); err != nil {
		log.Error("Failed Install:", err)
	}
	log.Info("Install success")

	// GetCertificates
	certs, err := client.GetCertificates(context.Background())
	if err != nil {
		log.Error("Failed GetCertificates:", err)
	}
	log.Info("GetCertificates success:\n", pretty.Sprint(certs))

	// // Rotate
	// params := pkix.Name{CommonName: "First Certificate"}
	// if err = client.Rotate(context.Background(), "hello certificate", params, sign); err != nil {
	// 	log.Error("Failed Install:", err)
	// }
	// log.Info("Install success")

	log.Info("Graceful exit.")
}
