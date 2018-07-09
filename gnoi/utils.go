package gnoi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
)

// FilterInternalPB returns true for protobuf internal variables in a path.
func FilterInternalPB(p cmp.Path) bool {
	return strings.Contains(p.String(), "XXX")
}

// Equal checks if two data structures are equal, ignoring protobuf internal variables.
func Equal(x, y interface{}) bool {
	return cmp.Equal(x, y, cmp.FilterPath(FilterInternalPB, cmp.Ignore()))
}

// Diff returns the diff between two data structures, ignoring protobuf internal variables.
func Diff(x, y interface{}) string {
	return cmp.Diff(x, y, cmp.FilterPath(FilterInternalPB, cmp.Ignore()))
}

// EncodeCert encodes a x509.Certificate into a PEM block.
func EncodeCert(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

// DecodeCert decodes a PEM block into a x509.Certificate.
func DecodeCert(bytes []byte) (*x509.Certificate, error) {
	certDERBlock, _ := pem.Decode(bytes)
	if certDERBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	certificate, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode DER bytes")
	}
	return certificate, nil
}

// DecodeCerts decodes a slice of PEM blocks into a slice of x509.Certificate.
func DecodeCerts(certs [][]byte) ([]*x509.Certificate, error) {
	res := []*x509.Certificate{}
	for _, cert := range certs {
		decoded, err := DecodeCert(cert)
		if err != nil {
			return nil, err
		}
		res = append(res, decoded)
	}
	return res, nil
}

// CreateSelfSignedCert returns a self signed certificate given a private key.
func CreateSelfSignedCert(privateKey crypto.PrivateKey) (*tls.Certificate, error) {
	subject := pkix.Name{
		Country:            []string{"NZ"},
		Organization:       []string{"gNxI"},
		OrganizationalUnit: []string{"gNxI"},
		Locality:           []string{"gNxI"},
		Province:           []string{"gNxI"},
		StreetAddress:      []string{"gNxI"},
		PostalCode:         []string{},
		CommonName:         "gNxI Server",
		Names:              []pkix.AttributeTypeAndValue{},
		ExtraNames:         []pkix.AttributeTypeAndValue{},
	}

	x509CertTemplate := &x509.Certificate{
		BasicConstraintsValid: true,
		DNSNames:              []string{},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		MaxPathLen:            5,
		IsCA:                  true,
		MaxPathLenZero:        true,
		NotAfter:              time.Now().Add(24 * 365 * time.Hour),
		NotBefore:             time.Now(),
		SignatureAlgorithm:    x509.SHA256WithRSA,
		Subject:               subject,
	}

	rsaPrivateKey, ok := (privateKey).(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse private key, not a rsa.PrivateKey type")
	}
	publicKey, ok := (rsaPrivateKey.Public()).(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key, not a rsa.PublicKey type")
	}
	pkBytes, err := asn1.Marshal(*publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}
	subjectKeyID := sha1.Sum(pkBytes)
	x509CertTemplate.SubjectKeyId = subjectKeyID[:]

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to randomize a big int: %v", err)
	}
	x509CertTemplate.SerialNumber = serialNumber

	derCert, err := x509.CreateCertificate(rand.Reader, x509CertTemplate, x509CertTemplate, publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	tlsCert := &tls.Certificate{}
	tlsCert.Leaf, err = x509.ParseCertificate(derCert)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the certificate: %v", err)
	}
	tlsCert.Certificate = [][]byte{tlsCert.Leaf.Raw}
	tlsCert.PrivateKey = privateKey

	return tlsCert, nil
}
