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

// Package entity provides a lightweight method for generating certificates.
package entity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

var (
	bigInt         = big.NewInt(0).Lsh(big.NewInt(1), 128)
	rsaBitSize     = 2048
	randReader     = rand.Reader
	certMaxPathLen = 5
	certExpiration = (365 * 24 * time.Hour)
)

// CreateSelfSigned creates an Entity with a self signed certificate.
func CreateSelfSigned(cn string, priv crypto.PrivateKey) (*Entity, error) {
	ca, err := NewEntity(TemplateCA(cn), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new Entity: %v", err)
	}
	if err := ca.SignWith(ca); err != nil {
		return nil, fmt.Errorf("failed to sign Entity: %v", err)
	}
	return ca, nil
}

// CreateSignedCA creates an Entity with a CA certificate signed by parent.
func CreateSignedCA(cn string, priv crypto.PrivateKey, parent *Entity) (*Entity, error) {
	ca, err := NewEntity(TemplateCA(cn), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new Entity: %v", err)
	}
	if err := ca.SignWith(parent); err != nil {
		return nil, fmt.Errorf("failed to sign Entity: %v", err)
	}
	return ca, nil
}

// CreateSigned creates an Entity with a certificate signed by parent.
func CreateSigned(cn string, priv crypto.PrivateKey, parent *Entity) (*Entity, error) {
	ca, err := NewEntity(Template(cn), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new Entity: %v", err)
	}
	if err := ca.SignWith(parent); err != nil {
		return nil, fmt.Errorf("failed to sign Entity: %v", err)
	}
	return ca, nil
}

// Entity contains a certificate, associated template, public and private keys.
type Entity struct {
	Template    *x509.Certificate
	PrivateKey  crypto.PrivateKey
	PublicKey   crypto.PublicKey
	Certificate *tls.Certificate
}

// FromFile loads an Entity with a certificate and private key from file.
func FromFile(certFile, privKeyFile string) (*Entity, error) {
	cert, err := tls.LoadX509KeyPair(certFile, privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load x509 key pair: %v", err)
	}
	if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return &Entity{PrivateKey: cert.PrivateKey, Certificate: &cert}, nil
}

// FromSigningRequest creates the boilerplate for a new certificate out of a Signing Request.
func FromSigningRequest(csr *x509.CertificateRequest) (*Entity, error) {
	template := &x509.Certificate{
		BasicConstraintsValid: true,
		DNSNames:              csr.DNSNames,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(certExpiration),
		NotBefore:             time.Now(),
		SignatureAlgorithm:    csr.SignatureAlgorithm,
		Subject:               csr.Subject,
		Signature:             csr.Signature,
		Extensions:            csr.Extensions,
		Version:               csr.Version,
		ExtraExtensions:       csr.ExtraExtensions,
		EmailAddresses:        csr.EmailAddresses,
		IPAddresses:           csr.IPAddresses,
		// URIs:                  csr.URIs,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
	}
	var err error
	if template.SubjectKeyId, err = keyID(csr.PublicKey); err != nil {
		return nil, fmt.Errorf("failed to generate Subject Key ID: %v", err)
	}
	if template.SerialNumber, err = rand.Int(randReader, bigInt); err != nil {
		return nil, fmt.Errorf("failed to randomize a big int: %v", err)
	}
	return &Entity{Template: template, PublicKey: csr.PublicKey}, nil
}

// NewEntity creates the boilerplate for a new certificate out of a template.
func NewEntity(template *x509.Certificate, privateKey crypto.PrivateKey) (*Entity, error) {
	priv, err := rsa.GenerateKey(randReader, rsaBitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %v", err)
	}
	if privateKey != nil {
		priv = privateKey.(*rsa.PrivateKey)
	}
	if template.SubjectKeyId, err = keyID(priv.Public()); err != nil {
		return nil, fmt.Errorf("failed to generate Subject Key ID: %v", err)
	}

	if template.SerialNumber, err = rand.Int(randReader, bigInt); err != nil {
		return nil, fmt.Errorf("failed to randomize a big int: %v", err)
	}

	return &Entity{Template: template, PrivateKey: priv, PublicKey: priv.Public(), Certificate: nil}, nil
}

func keyID(pub crypto.PublicKey) ([]byte, error) {
	pk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key, not a rsa.PublicKey type")
	}
	pkBytes, err := asn1.Marshal(*pk)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}
	subjectKeyID := sha1.Sum(pkBytes)
	return subjectKeyID[:], nil
}

// SignWith signs the boilerplate certificate with the parent certificate.
func (e *Entity) SignWith(parent *Entity) error {
	parentTemplate := parent.Template
	if parent.Certificate != nil && parent.Certificate.Leaf != nil {
		parentTemplate = parent.Certificate.Leaf
	}
	if parentTemplate == nil {
		return fmt.Errorf("no template found for signing the certificate")
	}

	e.Template.Issuer = parentTemplate.Subject
	e.Template.AuthorityKeyId = parentTemplate.SubjectKeyId
	derCert, err := x509.CreateCertificate(randReader, e.Template, parentTemplate, e.PublicKey, parent.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	tlsCert := &tls.Certificate{}
	tlsCert.Leaf, err = x509.ParseCertificate(derCert)
	if err != nil {
		return fmt.Errorf("failed to parse the certificate: %v", err)
	}
	tlsCert.Certificate = [][]byte{tlsCert.Leaf.Raw}
	tlsCert.PrivateKey = e.PrivateKey
	e.Certificate = tlsCert

	return nil
}

// SigningRequest generates a Certificate Signing Request out of the Entity.
func (e *Entity) SigningRequest() ([]byte, error) {
	csr := &x509.CertificateRequest{
		Attributes:      []pkix.AttributeTypeAndValueSET{},
		DNSNames:        e.Template.DNSNames,
		EmailAddresses:  e.Template.EmailAddresses,
		ExtraExtensions: e.Template.ExtraExtensions,
		IPAddresses:     e.Template.IPAddresses,
		// URIs:               e.Template.URIs,
		SignatureAlgorithm: e.Template.SignatureAlgorithm,
		Subject:            e.Template.Subject,
	}
	return x509.CreateCertificateRequest(randReader, csr, e.PrivateKey)
}

// SignedBy returns error if the certificate is not signed by parent.
func (e *Entity) SignedBy(parent *Entity) error {
	return e.Certificate.Leaf.CheckSignatureFrom(parent.Certificate.Leaf)
}

// TemplateCA returns a CA x509 template with cn as common name.
func TemplateCA(cn string) *x509.Certificate {
	ca := Template(cn)
	ca.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	ca.IsCA = true
	ca.MaxPathLen = certMaxPathLen
	ca.MaxPathLenZero = true
	return ca
}

// Template returns a leaf x509 template with cn as common name.
func Template(cn string) *x509.Certificate {
	subject := pkix.Name{
		Country:            []string{"NZ"},
		Organization:       []string{"gNxI"},
		OrganizationalUnit: []string{"gNxI"},
		Locality:           []string{"gNxI"},
		Province:           []string{"gNxI"},
		StreetAddress:      []string{"gNxI"},
		PostalCode:         []string{},
		CommonName:         cn,
		Names:              []pkix.AttributeTypeAndValue{},
		ExtraNames:         []pkix.AttributeTypeAndValue{},
	}

	return &x509.Certificate{
		// AuthorityKeyId,
		BasicConstraintsValid: true,
		DNSNames:              []string{},
		// ExcludedDNSDomains,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		// IsCA,
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:       time.Now().Add(24 * 365 * time.Hour),
		NotBefore:      time.Now(),
		// PermittedDNSDomains,
		// PermittedDNSDomainsCritical,
		// SerialNumber,
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject:            subject,
		// SubjectKeyId,
		// UnknownExtKeyUsage,
	}
}
