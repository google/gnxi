package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

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
func x509toPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

// DecodeCert decodes a PEM block into a x509.Certificate.
func PEMtox509(bytes []byte) (*x509.Certificate, error) {
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
