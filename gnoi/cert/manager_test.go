package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	cmpOpts = []cmp.Option{cmpopts.IgnoreUnexported(sync.RWMutex{}), cmp.AllowUnexported(Manager{}), cmpopts.IgnoreUnexported(CertInfo{}), cmpopts.IgnoreUnexported(x509.Certificate{})} //, cmpopts.IgnoreTypes(&x509.Certificate{})}
	now     = time.Now()
)

func TestRotate(t *testing.T) {
	originalDecoder := certPEMDecoder
	defer func() { certPEMDecoder = originalDecoder }()
	certPEMDecoder = func([]byte) (*x509.Certificate, error) { return &x509.Certificate{}, nil }

	tests := []struct {
		mgr        *Manager
		wantMgr    *Manager
		certID     string
		pemCert    []byte
		pemCACerts [][]byte
		wantError  bool
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id3",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id1",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  false,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id2",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
	}
	for _, test := range tests {
		_, rollback, err := test.mgr.Rotate(test.certID, test.pemCert, test.pemCACerts)
		if rollback != nil {
			rollback()
		}
		if err != nil && !test.wantError {
			t.Errorf("GetCertInfo error: %s", err)
		}
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
	}
}

func TestInstall(t *testing.T) {
	originalDecoder := certPEMDecoder
	defer func() { certPEMDecoder = originalDecoder }()
	certPEMDecoder = func([]byte) (*x509.Certificate, error) { return &x509.Certificate{}, nil }

	tests := []struct {
		mgr        *Manager
		wantMgr    *Manager
		certID     string
		pemCert    []byte
		pemCACerts [][]byte
		wantError  bool
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
					"id3": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}, &x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id3",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  false,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id4": &CertInfo{},
					"id5": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id5": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id4": &CertInfo{},
					"id5": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id5": true},
			},
			certID:     "id4",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
	}
	for _, test := range tests {
		err := test.mgr.Install(test.certID, test.pemCert, test.pemCACerts)
		if err != nil && !test.wantError {
			t.Errorf("GetCertInfo error: %s", err)
		}
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
	}
}

func TestGenCSR(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	tests := []struct {
		mgr     *Manager
		wantErr bool
	}{
		{
			mgr:     &Manager{},
			wantErr: true,
		},
		{
			mgr:     &Manager{privateKey: privKey},
			wantErr: false,
		},
	}
	for _, test := range tests {
		_, err := test.mgr.GenCSR(pkix.Name{})
		if err != nil && !test.wantErr {
			t.Errorf("GenCSR error: %s", err)
		}
	}
}

func TestGetCertInfo(t *testing.T) {
	tests := []struct {
		mgr          *Manager
		wantCertInfo []*CertInfo
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantCertInfo: []*CertInfo{
				&CertInfo{},
				&CertInfo{},
			},
		},
	}

	for _, test := range tests {
		gotCertInfo, err := test.mgr.GetCertInfo()
		if err != nil {
			t.Errorf("GetCertInfo error: %s", err)
		}
		if !cmp.Equal(test.wantCertInfo, gotCertInfo, cmpOpts...) {
			t.Errorf("GetCertInfo: (-want +got):\n%s", cmp.Diff(test.wantCertInfo, gotCertInfo, cmpOpts...))
		}
	}
}

func TestRevoke(t *testing.T) {
	tests := []struct {
		mgr     *Manager
		wantMgr *Manager
		in      []string
		wantRvk []string
		wantErr map[string]string
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id1": &CertInfo{},
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*CertInfo{
					"id2": &CertInfo{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			in:      []string{"id1", "id2", "id3"},
			wantRvk: []string{"id1"},
			wantErr: map[string]string{
				"id2": "an operation with this certID is in progress",
				"id3": "does not exist",
			},
		},
	}

	for _, test := range tests {
		gotRvk, gotErr, _ := test.mgr.Revoke(test.in)
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Revoke: Manager (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
		if !cmp.Equal(test.wantRvk, gotRvk) {
			t.Errorf("Revoke: revoked (-want +got):\n%s", cmp.Diff(test.wantRvk, gotRvk))
		}
		if !cmp.Equal(test.wantErr, gotErr) {
			t.Errorf("Revoke: Manager (-want +got):\n%s", cmp.Diff(test.wantErr, gotErr))
		}
	}
}

// func noTestCertManager(t *testing.T) {
// 	privateKey := "some random key"
// 	now := time.Now()
// 	nowTime = func() time.Time { return now }
// 	cmpOpts := []cmp.Option{cmpopts.IgnoreUnexported(sync.RWMutex{}), cmp.AllowUnexported(Manager{})}
// 	expectCert1, err := DecodeCert([]byte(exampleCertPEM1))
// 	if err != nil {
// 		t.Fatal("failed DecodeCert:", err)
// 	}
// 	expectCert2, err := DecodeCert([]byte(exampleCertPEM2))
// 	if err != nil {
// 		t.Fatal("failed DecodeCert:", err)
// 	}
// 	certID1 := "id1"
// 	certID2 := "id2"
// 	certID3 := "id3"
//
// 	t.Run(("Install new certID"), func(t *testing.T) {
// 		cm := NewManager(privateKey)
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert1,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
// 		}
// 		want := &Manager{
// 			privateKey:   privateKey,
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{},
// 			notifiers:    []Notifier{},
// 		}
// 		if err := cm.Install(param); err != nil {
// 			t.Fatal("failed Install:", err)
// 		}
// 		if !cmp.Equal(want, cm, cmpOpts...) {
// 			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Install existing certID"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{},
// 		}
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert1,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
// 		}
// 		if err := cm.Install(param); err == nil {
// 			t.Errorf("expected failed Install")
// 		}
// 	})
//
// 	t.Run(("Install on changing certID"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID1: true},
// 		}
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert1,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
// 		}
// 		if err := cm.Install(param); err == nil {
// 			t.Errorf("expected failed Install: %v", err)
// 		}
// 	})
//
// 	t.Run(("Get"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID1: true},
// 		}
// 		want := []*pb.CertificateInfo{
// 			&pb.CertificateInfo{
// 				CertificateId: certID1,
// 				Certificate: &pb.Certificate{
// 					Type:        pb.CertificateType_CT_X509,
// 					Certificate: EncodeCert(expectCert1),
// 				},
// 				ModificationTime: now.UnixNano(),
// 			},
// 		}
// 		got, err := cm.Get()
// 		if err != nil {
// 			t.Fatal("failed Get:", err)
// 		}
// 		if !cmp.Equal(want, got, cmpOpts...) {
// 			t.Errorf("Get: (-want +got):\n%s", cmp.Diff(want, got, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Rotate existing certID with Accept"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{},
// 		}
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert2,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert2, examplePbCert2, examplePbCert2},
// 		}
// 		rotateAccept, _, err := cm.Rotate(param)
// 		if err != nil {
// 			t.Fatal("failed Rotate:", err)
// 		}
//
// 		want := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert2},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
// 			locks:        map[string]bool{certID1: true},
// 		}
// 		if !cmp.Equal(want, cm, cmpOpts...) {
// 			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
// 		}
// 		rotateAccept()
// 		want = &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert2},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
// 			locks:        map[string]bool{},
// 		}
// 		if !cmp.Equal(want, cm, cmpOpts...) {
// 			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Rotate existing certID with Rollback"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{},
// 		}
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert2,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert2, examplePbCert2, examplePbCert2},
// 		}
// 		_, rotateBack, err := cm.Rotate(param)
// 		if err != nil {
// 			t.Fatal("failed Rotate:", err)
// 		}
//
// 		want := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert2},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
// 			locks:        map[string]bool{certID1: true},
// 		}
// 		if !cmp.Equal(want, cm, cmpOpts...) {
// 			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
// 		}
// 		rotateBack()
// 		want = &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{},
// 		}
// 		if !cmp.Equal(want, cm, cmpOpts...) {
// 			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Rotate on changing certID"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID1: true},
// 		}
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert1,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
// 		}
// 		if _, _, err := cm.Rotate(param); err == nil {
// 			t.Errorf("expected failed Rotate: %v", err)
// 		}
// 	})
//
// 	t.Run(("Rotate unexisting cerID"), func(t *testing.T) {
// 		cm := NewManager(privateKey)
// 		param := &pb.LoadCertificateRequest{
// 			Certificate:   examplePbCert1,
// 			CertificateId: certID1,
// 			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
// 		}
// 		if _, _, err := cm.Rotate(param); err == nil {
// 			t.Errorf("expected failed Rotate: %v", err)
// 		}
// 	})
//
// 	t.Run(("Revoke"), func(t *testing.T) {
// 		cm := &Manager{
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID3: true},
// 		}
// 		param := &pb.RevokeCertificatesRequest{CertificateId: []string{certID1, certID2, certID3}}
// 		wantCM := &Manager{
// 			certs:        map[string]*x509.Certificate{},
// 			certsModTime: map[string]time.Time{},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID3: true},
// 		}
// 		wantResponse := &pb.RevokeCertificatesResponse{
// 			RevokedCertificateId: []string{certID1},
// 			CertificateRevocationError: []*pb.CertificateRevocationError{
// 				&pb.CertificateRevocationError{
// 					CertificateId: certID2,
// 					ErrorMessage:  "does not exist",
// 				},
// 				&pb.CertificateRevocationError{
// 					CertificateId: certID3,
// 					ErrorMessage:  "an operation with this certID is in progress",
// 				},
// 			},
// 		}
// 		gotResp, err := cm.Revoke(param)
// 		if err != nil {
// 			t.Fatal("failed Revoke:", err)
// 		}
// 		if !cmp.Equal(wantResponse, gotResp, cmpOpts...) {
// 			t.Errorf("Rotate response: (-want +got):\n%s", cmp.Diff(wantResponse, gotResp, cmpOpts...))
// 		}
// 		if !cmp.Equal(wantCM, cm, cmpOpts...) {
// 			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(wantCM, cm, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Certificates"), func(t *testing.T) {
// 		cm := &Manager{
// 			privateKey:   privateKey,
// 			certs:        map[string]*x509.Certificate{certID1: expectCert1},
// 			certsModTime: map[string]time.Time{certID1: now},
// 			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
// 			locks:        map[string]bool{certID3: true},
// 		}
// 		wantTLSCerts := []tls.Certificate{
// 			tls.Certificate{Leaf: expectCert1, Certificate: [][]byte{expectCert1.Raw}, PrivateKey: privateKey},
// 		}
// 		gotTLSCerts, _ := cm.Certificates()
// 		if !cmp.Equal(wantTLSCerts, gotTLSCerts, cmpOpts...) {
// 			t.Errorf("TLS Certificates: (-want +got):\n%s", cmp.Diff(wantTLSCerts, gotTLSCerts, cmpOpts...))
// 		}
// 	})
//
// 	t.Run(("Empty"), func(t *testing.T) {
// 		tests := []struct {
// 			cm   *Manager
// 			want bool
// 		}{
// 			{
// 				cm: &Manager{
// 					certs:        map[string]*x509.Certificate{},
// 					certsModTime: map[string]time.Time{},
// 					caBundle:     []*x509.Certificate{},
// 					locks:        map[string]bool{},
// 				},
// 				want: true,
// 			},
// 			{
// 				cm: &Manager{
// 					certs:    map[string]*x509.Certificate{certID1: expectCert1},
// 					caBundle: []*x509.Certificate{},
// 					locks:    map[string]bool{},
// 				},
// 				want: false,
// 			},
// 			{
// 				cm: &Manager{
// 					certs:    map[string]*x509.Certificate{},
// 					caBundle: []*x509.Certificate{},
// 					locks:    map[string]bool{certID3: true},
// 				},
// 				want: false,
// 			},
// 			{
// 				cm: &Manager{
// 					certs:    map[string]*x509.Certificate{certID1: expectCert1},
// 					caBundle: []*x509.Certificate{expectCert1},
// 					locks:    map[string]bool{},
// 				},
// 				want: false,
// 			},
// 		}
// 		for _, test := range tests {
// 			if test.cm.Empty() != test.want {
// 				t.Errorf("Expected %v but got %v for:\nlen(certs): %d\nlen(caBundle): %d\nlen(locks): %d", test.want, !test.want, len(test.cm.certs), len(test.cm.caBundle), len(test.cm.locks))
// 			}
// 		}
// 	})
//
// 	cm := NewManager(privateKey)
// 	createCSR = func(rand io.Reader, template *x509.CertificateRequest, priv interface{}) (csr []byte, err error) {
// 		return []byte("hello"), nil
// 	}
// 	t.Run(("GenCSR unsupported certificate type"), func(t *testing.T) {
// 		_, err := cm.GenCSR(&pb.CSRParams{KeyType: pb.KeyType_KT_RSA})
// 		if err == nil {
// 			t.Errorf("expected failed GenCSR")
// 		}
// 	})
//
// 	t.Run(("GenCSR unsupported key type"), func(t *testing.T) {
// 		_, err := cm.GenCSR(&pb.CSRParams{Type: pb.CertificateType_CT_X509})
// 		if err == nil {
// 			t.Errorf("expected failed GenCSR")
// 		}
// 	})
//
// 	t.Run(("GenCSR"), func(t *testing.T) {
// 		want := &pb.CSR{
// 			Type: pb.CertificateType_CT_X509,
// 			Csr:  []byte("hello"),
// 		}
// 		got, err := cm.GenCSR(&pb.CSRParams{Type: pb.CertificateType_CT_X509, KeyType: pb.KeyType_KT_RSA})
// 		if err != nil {
// 			t.Errorf("failed GenCSR: %v", err)
// 		}
// 		if !Equal(want, got) {
// 			t.Errorf("GenCSR: (-want +got):\n%s", Diff(want, got))
// 		}
// 	})
// }
