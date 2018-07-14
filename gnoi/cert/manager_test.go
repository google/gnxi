package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/gnxi/gnoi/cert/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCertManager(t *testing.T) {
	privateKey := "some random key"
	now := time.Now()
	nowTime = func() time.Time { return now }
	cmpOpts := []cmp.Option{cmpopts.IgnoreUnexported(sync.RWMutex{}), cmp.AllowUnexported(Manager{})}
	expectCert1, err := DecodeCert([]byte(exampleCertPEM1))
	if err != nil {
		t.Fatal("failed DecodeCert:", err)
	}
	expectCert2, err := DecodeCert([]byte(exampleCertPEM2))
	if err != nil {
		t.Fatal("failed DecodeCert:", err)
	}
	certID1 := "id1"
	certID2 := "id2"
	certID3 := "id3"

	t.Run(("Install new certID"), func(t *testing.T) {
		cm := NewManager(privateKey)
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		want := &Manager{
			privateKey:   privateKey,
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
			notifiers:    []Notifier{},
		}
		if err := cm.Install(param); err != nil {
			t.Fatal("failed Install:", err)
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
	})

	t.Run(("Install existing certID"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
		}
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		if err := cm.Install(param); err == nil {
			t.Errorf("expected failed Install")
		}
	})

	t.Run(("Install on changing certID"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID1: true},
		}
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		if err := cm.Install(param); err == nil {
			t.Errorf("expected failed Install: %v", err)
		}
	})

	t.Run(("Get"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID1: true},
		}
		want := []*pb.CertificateInfo{
			&pb.CertificateInfo{
				CertificateId: certID1,
				Certificate: &pb.Certificate{
					Type:        pb.CertificateType_CT_X509,
					Certificate: EncodeCert(expectCert1),
				},
				ModificationTime: now.UnixNano(),
			},
		}
		got, err := cm.Get()
		if err != nil {
			t.Fatal("failed Get:", err)
		}
		if !cmp.Equal(want, got, cmpOpts...) {
			t.Errorf("Get: (-want +got):\n%s", cmp.Diff(want, got, cmpOpts...))
		}
	})

	t.Run(("Rotate existing certID with Accept"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
		}
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert2,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert2, examplePbCert2, examplePbCert2},
		}
		rotateAccept, _, err := cm.Rotate(param)
		if err != nil {
			t.Fatal("failed Rotate:", err)
		}

		want := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert2},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
			locks:        map[string]bool{certID1: true},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
		rotateAccept()
		want = &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert2},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
			locks:        map[string]bool{},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
	})

	t.Run(("Rotate existing certID with Rollback"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
		}
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert2,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert2, examplePbCert2, examplePbCert2},
		}
		_, rotateBack, err := cm.Rotate(param)
		if err != nil {
			t.Fatal("failed Rotate:", err)
		}

		want := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert2},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
			locks:        map[string]bool{certID1: true},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
		rotateBack()
		want = &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
	})

	t.Run(("Rotate on changing certID"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID1: true},
		}
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		if _, _, err := cm.Rotate(param); err == nil {
			t.Errorf("expected failed Rotate: %v", err)
		}
	})

	t.Run(("Rotate unexisting cerID"), func(t *testing.T) {
		cm := NewManager(privateKey)
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		if _, _, err := cm.Rotate(param); err == nil {
			t.Errorf("expected failed Rotate: %v", err)
		}
	})

	t.Run(("Revoke"), func(t *testing.T) {
		cm := &Manager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID3: true},
		}
		param := &pb.RevokeCertificatesRequest{CertificateId: []string{certID1, certID2, certID3}}
		wantCM := &Manager{
			certs:        map[string]*x509.Certificate{},
			certsModTime: map[string]time.Time{},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID3: true},
		}
		wantResponse := &pb.RevokeCertificatesResponse{
			RevokedCertificateId: []string{certID1},
			CertificateRevocationError: []*pb.CertificateRevocationError{
				&pb.CertificateRevocationError{
					CertificateId: certID2,
					ErrorMessage:  "does not exist",
				},
				&pb.CertificateRevocationError{
					CertificateId: certID3,
					ErrorMessage:  "an operation with this certID is in progress",
				},
			},
		}
		gotResp, err := cm.Revoke(param)
		if err != nil {
			t.Fatal("failed Revoke:", err)
		}
		if !cmp.Equal(wantResponse, gotResp, cmpOpts...) {
			t.Errorf("Rotate response: (-want +got):\n%s", cmp.Diff(wantResponse, gotResp, cmpOpts...))
		}
		if !cmp.Equal(wantCM, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(wantCM, cm, cmpOpts...))
		}
	})

	t.Run(("Certificates"), func(t *testing.T) {
		cm := &Manager{
			privateKey:   privateKey,
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID3: true},
		}
		wantTLSCerts := []tls.Certificate{
			tls.Certificate{Leaf: expectCert1, Certificate: [][]byte{expectCert1.Raw}, PrivateKey: privateKey},
		}
		gotTLSCerts, _ := cm.Certificates()
		if !cmp.Equal(wantTLSCerts, gotTLSCerts, cmpOpts...) {
			t.Errorf("TLS Certificates: (-want +got):\n%s", cmp.Diff(wantTLSCerts, gotTLSCerts, cmpOpts...))
		}
	})

	t.Run(("Empty"), func(t *testing.T) {
		tests := []struct {
			cm   *Manager
			want bool
		}{
			{
				cm: &Manager{
					certs:        map[string]*x509.Certificate{},
					certsModTime: map[string]time.Time{},
					caBundle:     []*x509.Certificate{},
					locks:        map[string]bool{},
				},
				want: true,
			},
			{
				cm: &Manager{
					certs:    map[string]*x509.Certificate{certID1: expectCert1},
					caBundle: []*x509.Certificate{},
					locks:    map[string]bool{},
				},
				want: false,
			},
			{
				cm: &Manager{
					certs:    map[string]*x509.Certificate{},
					caBundle: []*x509.Certificate{},
					locks:    map[string]bool{certID3: true},
				},
				want: false,
			},
			{
				cm: &Manager{
					certs:    map[string]*x509.Certificate{certID1: expectCert1},
					caBundle: []*x509.Certificate{expectCert1},
					locks:    map[string]bool{},
				},
				want: false,
			},
		}
		for _, test := range tests {
			if test.cm.Empty() != test.want {
				t.Errorf("Expected %v but got %v for:\nlen(certs): %d\nlen(caBundle): %d\nlen(locks): %d", test.want, !test.want, len(test.cm.certs), len(test.cm.caBundle), len(test.cm.locks))
			}
		}
	})

	cm := NewManager(privateKey)
	createCSR = func(rand io.Reader, template *x509.CertificateRequest, priv interface{}) (csr []byte, err error) {
		return []byte("hello"), nil
	}
	t.Run(("GenCSR unsupported certificate type"), func(t *testing.T) {
		_, err := cm.GenCSR(&pb.CSRParams{KeyType: pb.KeyType_KT_RSA})
		if err == nil {
			t.Errorf("expected failed GenCSR")
		}
	})

	t.Run(("GenCSR unsupported key type"), func(t *testing.T) {
		_, err := cm.GenCSR(&pb.CSRParams{Type: pb.CertificateType_CT_X509})
		if err == nil {
			t.Errorf("expected failed GenCSR")
		}
	})

	t.Run(("GenCSR"), func(t *testing.T) {
		want := &pb.CSR{
			Type: pb.CertificateType_CT_X509,
			Csr:  []byte("hello"),
		}
		got, err := cm.GenCSR(&pb.CSRParams{Type: pb.CertificateType_CT_X509, KeyType: pb.KeyType_KT_RSA})
		if err != nil {
			t.Errorf("failed GenCSR: %v", err)
		}
		if !Equal(want, got) {
			t.Errorf("GenCSR: (-want +got):\n%s", Diff(want, got))
		}
	})
}
