package gnoi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/google/gnxi/gnoi/certpb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	expectCSRParams = &pb.CSRParams{
		Type:               pb.CertificateType_CT_X509,
		KeyType:            pb.KeyType_KT_RSA,
		MinKeySize:         2048,
		CommonName:         "gNXI",
		Country:            "NZ",
		State:              "gNXI",
		City:               "gNXI",
		Organization:       "gNXI",
		OrganizationalUnit: "gNXI",
		IpAddress:          "1.2.3.4",
		EmailId:            "gNXI@gNXI",
	}
	expectCSR = &pb.CSR{
		Type: pb.CertificateType_CT_X509,
		Csr:  make([]byte, 11),
	}
	expectLoadCertificateRequest = &pb.LoadCertificateRequest{
		Certificate: &pb.Certificate{
			Type:        pb.CertificateType_CT_X509,
			Certificate: make([]byte, 22),
		},
		CertificateId: "id1",
		CaCertificate: []*pb.Certificate{
			&pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: make([]byte, 33),
			},
			&pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: make([]byte, 44),
			},
		},
	}
	expectCertificateInfo = []*pb.CertificateInfo{
		&pb.CertificateInfo{
			CertificateId: "id1",
			Certificate: &pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: make([]byte, 55),
			},
			Endpoints:        nil,
			ModificationTime: time.Now().Unix(),
		},
		&pb.CertificateInfo{
			CertificateId: "id2",
			Certificate: &pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: make([]byte, 66),
			},
			Endpoints:        nil,
			ModificationTime: time.Now().Unix(),
		},
	}
	expectRevokeCertificatesRequest = &pb.RevokeCertificatesRequest{
		CertificateId: []string{"id1", "id2"},
	}
	expectRevokeCertificatesResponse = &pb.RevokeCertificatesResponse{
		RevokedCertificateId: []string{"id1", "id2"},
		CertificateRevocationError: []*pb.CertificateRevocationError{
			&pb.CertificateRevocationError{CertificateId: "id3", ErrorMessage: "error msg 3"},
			&pb.CertificateRevocationError{CertificateId: "id4", ErrorMessage: "error msg 4"},
		},
	}
	exampleCertPEM1 = `
-----BEGIN CERTIFICATE-----
MIIEBDCCAuygAwIBAgIDAjppMA0GCSqGSIb3DQEBBQUAMEIxCzAJBgNVBAYTAlVT
MRYwFAYDVQQKEw1HZW9UcnVzdCBJbmMuMRswGQYDVQQDExJHZW9UcnVzdCBHbG9i
YWwgQ0EwHhcNMTMwNDA1MTUxNTU1WhcNMTUwNDA0MTUxNTU1WjBJMQswCQYDVQQG
EwJVUzETMBEGA1UEChMKR29vZ2xlIEluYzElMCMGA1UEAxMcR29vZ2xlIEludGVy
bmV0IEF1dGhvcml0eSBHMjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AJwqBHdc2FCROgajguDYUEi8iT/xGXAaiEZ+4I/F8YnOIe5a/mENtzJEiaB0C1NP
VaTOgmKV7utZX8bhBYASxF6UP7xbSDj0U/ck5vuR6RXEz/RTDfRK/J9U3n2+oGtv
h8DQUB8oMANA2ghzUWx//zo8pzcGjr1LEQTrfSTe5vn8MXH7lNVg8y5Kr0LSy+rE
ahqyzFPdFUuLH8gZYR/Nnag+YyuENWllhMgZxUYi+FOVvuOAShDGKuy6lyARxzmZ
EASg8GF6lSWMTlJ14rbtCMoU/M4iarNOz0YDl5cDfsCx3nuvRTPPuj5xt970JSXC
DTWJnZ37DhF5iR43xa+OcmkCAwEAAaOB+zCB+DAfBgNVHSMEGDAWgBTAephojYn7
qwVkDBF9qn1luMrMTjAdBgNVHQ4EFgQUSt0GFhu89mi1dvWBtrtiGrpagS8wEgYD
VR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAQYwOgYDVR0fBDMwMTAvoC2g
K4YpaHR0cDovL2NybC5nZW90cnVzdC5jb20vY3Jscy9ndGdsb2JhbC5jcmwwPQYI
KwYBBQUHAQEEMTAvMC0GCCsGAQUFBzABhiFodHRwOi8vZ3RnbG9iYWwtb2NzcC5n
ZW90cnVzdC5jb20wFwYDVR0gBBAwDjAMBgorBgEEAdZ5AgUBMA0GCSqGSIb3DQEB
BQUAA4IBAQA21waAESetKhSbOHezI6B1WLuxfoNCunLaHtiONgaX4PCVOzf9G0JY
/iLIa704XtE7JW4S615ndkZAkNoUyHgN7ZVm2o6Gb4ChulYylYbc3GrKBIxbf/a/
zG+FA1jDaFETzf3I93k9mTXwVqO94FntT0QJo544evZG0R0SnU++0ED8Vf4GXjza
HFa9llF7b1cq26KqltyMdMKVvvBulRP/F/A8rLIQjcxz++iPAsbw+zOzlTvjwsto
WHPbqCRiOwY1nQ2pM714A5AuTHhdUDqB1O6gyHA43LL5Z/qHQF1hwFGPa4NrzQU6
yuGnBXj8ytqU0CwIPX4WecigUCAkVDNx
-----END CERTIFICATE-----`
	exampleCertPEM2 = `
-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIE31FZVaPXTUwDQYJKoZIhvcNAQEFBQAwSTELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkdvb2dsZSBJbmMxJTAjBgNVBAMTHEdvb2dsZSBJbnRl
cm5ldCBBdXRob3JpdHkgRzIwHhcNMTQwMTI5MTMyNzQzWhcNMTQwNTI5MDAwMDAw
WjBpMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwN
TW91bnRhaW4gVmlldzETMBEGA1UECgwKR29vZ2xlIEluYzEYMBYGA1UEAwwPbWFp
bC5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfRrObuSW5T7q
5CnSEqefEmtH4CCv6+5EckuriNr1CjfVvqzwfAhopXkLrq45EQm8vkmf7W96XJhC
7ZM0dYi1/qOCAU8wggFLMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAa
BgNVHREEEzARgg9tYWlsLmdvb2dsZS5jb20wCwYDVR0PBAQDAgeAMGgGCCsGAQUF
BwEBBFwwWjArBggrBgEFBQcwAoYfaHR0cDovL3BraS5nb29nbGUuY29tL0dJQUcy
LmNydDArBggrBgEFBQcwAYYfaHR0cDovL2NsaWVudHMxLmdvb2dsZS5jb20vb2Nz
cDAdBgNVHQ4EFgQUiJxtimAuTfwb+aUtBn5UYKreKvMwDAYDVR0TAQH/BAIwADAf
BgNVHSMEGDAWgBRK3QYWG7z2aLV29YG2u2IaulqBLzAXBgNVHSAEEDAOMAwGCisG
AQQB1nkCBQEwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL3BraS5nb29nbGUuY29t
L0dJQUcyLmNybDANBgkqhkiG9w0BAQUFAAOCAQEAH6RYHxHdcGpMpFE3oxDoFnP+
gtuBCHan2yE2GRbJ2Cw8Lw0MmuKqHlf9RSeYfd3BXeKkj1qO6TVKwCh+0HdZk283
TZZyzmEOyclm3UGFYe82P/iDFt+CeQ3NpmBg+GoaVCuWAARJN/KfglbLyyYygcQq
0SgeDh8dRKUiaW3HQSoYvTvdTuqzwK4CXsr3b5/dAOY8uMuG/IAR3FgwTbZ1dtoW
RvOTa8hYiU6A475WuZKyEHcwnGYe57u2I2KbMgcKjPniocj4QzgYsVAVKW3IwaOh
yE+vPxsiUkvQHdO2fojCkY8jg70jxM+gu59tPDNbw3Uh/2Ij310FgTHsnGQMyA==
-----END CERTIFICATE-----`
	examplePbCert1 = &pb.Certificate{
		Certificate: []byte(exampleCertPEM1),
		Type:        pb.CertificateType_CT_X509,
	}
	examplePbCert2 = &pb.Certificate{
		Certificate: []byte(exampleCertPEM2),
		Type:        pb.CertificateType_CT_X509,
	}
)

type mockCertInterface struct {
	CertInterface
}

func (m *mockCertInterface) GenCSR(p *pb.CSRParams) (*pb.CSR, error) {
	if !Equal(expectCSRParams, p) {
		return nil, fmt.Errorf("GenCSR: (-want +got):\n%s", Diff(expectCSRParams, p))
	}
	return expectCSR, nil
}

func (m *mockCertInterface) Get() ([]*pb.CertificateInfo, error) {
	return expectCertificateInfo, nil
}

func (m *mockCertInterface) Install(l *pb.LoadCertificateRequest) error {
	if !Equal(expectLoadCertificateRequest, l) {
		return fmt.Errorf("Install: (-want +got):\n%s", Diff(expectLoadCertificateRequest, l))
	}
	return nil
}

func (m *mockCertInterface) Revoke(r *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	if !Equal(expectRevokeCertificatesRequest, r) {
		return nil, fmt.Errorf("Revoke: (-want +got):\n%s", Diff(expectRevokeCertificatesRequest, r))
	}
	return expectRevokeCertificatesResponse, nil
}

func (m *mockCertInterface) Rotate(l *pb.LoadCertificateRequest) (func(), func(), error) {
	if !Equal(expectLoadCertificateRequest, l) {
		return nil, nil, fmt.Errorf("Rotate: (-want +got):\n%s", Diff(expectLoadCertificateRequest, l))
	}
	return func() {}, func() {}, nil
}

func TestCertServer(t *testing.T) {
	conString := "127.0.0.1:4455"
	gServer := grpc.NewServer()
	pb.RegisterCertificateManagementServer(gServer, NewCertServer(&mockCertInterface{}))
	listen, err := net.Listen("tcp", conString)
	if err != nil {
		t.Fatal("server failed to listen:", err)
	}
	go gServer.Serve(listen)
	defer gServer.GracefulStop()
	time.Sleep(time.Second)

	dial, err := grpc.Dial(conString, grpc.WithInsecure())
	defer dial.Close()
	if err != nil {
		t.Fatal("client failed to dial:", err)
	}
	gClient := pb.NewCertificateManagementClient(dial)
	ctx := context.Background()

	t.Run("Rotate Case 1", func(t *testing.T) {
		stream, err := gClient.Rotate(ctx)
		if err != nil {
			t.Errorf("failed stream: %v", err)
			return
		}
		if err = stream.Send(&pb.RotateCertificateRequest{
			RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{
				GenerateCsr: &pb.GenerateCSRRequest{
					CsrParams: expectCSRParams,
				},
			},
		}); err != nil {
			t.Errorf("failed to send GenerateCSRRequest: %v", err)
			return
		}
		var req *pb.RotateCertificateResponse
		if req, err = stream.Recv(); err != nil {
			t.Errorf("failed to receive RotateCertificateResponse: %v", err)
			return
		}
		genCSR := req.GetGeneratedCsr()
		if genCSR == nil {
			t.Errorf("expected GenerateCSRRequest, got something else")
			return
		}
		if !Equal(expectCSR, genCSR.Csr) {
			t.Errorf("GetGeneratedCsr: (-want +got):\n%s", Diff(expectCSR, genCSR.Csr))
			return
		}

		// sign genCSR.Csr with CA
		if err = stream.Send(&pb.RotateCertificateRequest{
			RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{
				LoadCertificate: expectLoadCertificateRequest,
			},
		}); err != nil {
			t.Errorf("failed to send LoadCertificateRequest: %v", err)
			return
		}
		if req, err = stream.Recv(); err != nil {
			t.Errorf("failed to receive RotateCertificateResponse: %v", err)
			return
		}
		loadCertificateResponse := req.GetLoadCertificate()
		if loadCertificateResponse == nil {
			t.Errorf("expected LoadCertificateResponse, got something else")
			return
		}

		// Verify here.
		if err := stream.Send(&pb.RotateCertificateRequest{
			RotateRequest: &pb.RotateCertificateRequest_FinalizeRotation{FinalizeRotation: &pb.FinalizeRequest{}},
		}); err != nil {
			t.Errorf("failed to send LoadCertificateRequest: %v", err)
			return
		}
	})

	t.Run("Install Case 1", func(t *testing.T) {
		stream, err := gClient.Install(ctx)
		if err != nil {
			t.Errorf("failed stream: %v", err)
			return
		}

		if err = stream.Send(&pb.InstallCertificateRequest{
			InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{
				GenerateCsr: &pb.GenerateCSRRequest{CsrParams: expectCSRParams},
			},
		}); err != nil {
			t.Errorf("failed to send GenerateCSRRequest: %v", err)
			return
		}

		var req *pb.InstallCertificateResponse
		if req, err = stream.Recv(); err != nil {
			t.Errorf("failed to receive InstallCertificateResponse: %v", err)
			return
		}

		genCSR := req.GetGeneratedCsr()
		if genCSR == nil {
			t.Errorf("expected GenerateCSRRequest, got something else")
			return
		}
		if !Equal(expectCSR, genCSR.Csr) {
			t.Errorf("GetGeneratedCsr: (-want +got):\n%s", Diff(expectCSR, genCSR.Csr))
			return
		}
		// sign genCSR.Csr with CA

		if err = stream.Send(&pb.InstallCertificateRequest{
			InstallRequest: &pb.InstallCertificateRequest_LoadCertificate{
				LoadCertificate: expectLoadCertificateRequest,
			},
		}); err != nil {
			t.Errorf("failed to send LoadCertificateRequest: %v", err)
			return
		}

		if req, err = stream.Recv(); err != nil {
			t.Errorf("failed to receive InstallCertificateResponse: %v", err)
			return
		}
		loadCertificateResponse := req.GetLoadCertificate()
		if loadCertificateResponse == nil {
			t.Errorf("expected LoadCertificateResponse, got something else")
			return
		}
	})

	t.Run("GetCertificates", func(t *testing.T) {
		resp, err := gClient.GetCertificates(ctx, &pb.GetCertificatesRequest{})
		if err != nil {
			t.Errorf("GetCertificates: %v", err)
		}
		if !Equal(expectCertificateInfo, resp.CertificateInfo) {
			t.Errorf("GetCertificates: (-want +got):\n%s", Diff(expectCertificateInfo, resp.CertificateInfo))
		}
	})

	t.Run("RevokeCertificates", func(t *testing.T) {
		resp, err := gClient.RevokeCertificates(ctx, expectRevokeCertificatesRequest)
		if err != nil {
			t.Errorf("RevokeCertificates: %v", err)
		}
		if !Equal(expectRevokeCertificatesResponse, resp) {
			t.Errorf("GetCertificates: (-want +got):\n%s", Diff(expectRevokeCertificatesResponse, resp))
		}
	})

	t.Run("CanGenerateCSR", func(t *testing.T) {
		result, err := gClient.CanGenerateCSR(ctx, &pb.CanGenerateCSRRequest{
			KeyType:         pb.KeyType_KT_RSA,
			CertificateType: pb.CertificateType_CT_X509,
			KeySize:         2048,
		})
		if !result.CanGenerate {
			t.Errorf("CanGenerateCSR cannot generate CSR")
		}
		if err != nil {
			t.Errorf("CanGenerateCSR: %v", err)
		}
	})
}

func TestCertManager(t *testing.T) {

	privateKey := "some random key"

	t.Run("Lock & Unlock", func(t *testing.T) {
		cm := NewCertManager(privateKey)
		certID := "test"
		if err := cm.lock(certID); err != nil {
			t.Errorf("failed to lock: %v", err)
		}
		if err := cm.lock(certID); err == nil {
			t.Errorf("allowed double lock")
		}
		if !cm.locked(certID) {
			t.Errorf("locked but not locked")
		}
		cm.unlock(certID)
		if err := cm.lock(certID); err != nil {
			t.Errorf("failed to lock: %v", err)
		}
	})

	now := time.Now()
	nowTime = func() time.Time { return now }
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreUnexported(sync.Mutex{}),
		// cmpopts.IgnoreUnexported(rsa.PrivateKey{}),
		// cmpopts.IgnoreUnexported(rsa.PublicKey{}),
		cmp.AllowUnexported(CertManager{}),
	}
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
		cm := NewCertManager(privateKey)
		param := &pb.LoadCertificateRequest{
			Certificate:   examplePbCert1,
			CertificateId: certID1,
			CaCertificate: []*pb.Certificate{examplePbCert1, examplePbCert1, examplePbCert1},
		}
		want := &CertManager{
			privateKey:   privateKey,
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{},
		}
		if err := cm.Install(param); err != nil {
			t.Fatal("failed Install:", err)
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
	})

	t.Run(("Install existing certID"), func(t *testing.T) {
		cm := &CertManager{
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
			t.Errorf("expected failed Install: %v", err)
		}
	})

	t.Run(("Install on changing certID"), func(t *testing.T) {
		cm := &CertManager{
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
		cm := &CertManager{
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
		cm := &CertManager{
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

		want := &CertManager{
			certs:        map[string]*x509.Certificate{certID1: expectCert2},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
			locks:        map[string]bool{certID1: true},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
		rotateAccept()
		want = &CertManager{
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
		cm := &CertManager{
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

		want := &CertManager{
			certs:        map[string]*x509.Certificate{certID1: expectCert2},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert2, expectCert2, expectCert2},
			locks:        map[string]bool{certID1: true},
		}
		if !cmp.Equal(want, cm, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(want, cm, cmpOpts...))
		}
		rotateBack()
		want = &CertManager{
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
		cm := &CertManager{
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
		cm := NewCertManager(privateKey)
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
		cm := &CertManager{
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID3: true},
		}
		param := &pb.RevokeCertificatesRequest{CertificateId: []string{certID1, certID2, certID3}}
		wantCM := &CertManager{
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
		cm := &CertManager{
			privateKey:   privateKey,
			certs:        map[string]*x509.Certificate{certID1: expectCert1},
			certsModTime: map[string]time.Time{certID1: now},
			caBundle:     []*x509.Certificate{expectCert1, expectCert1, expectCert1},
			locks:        map[string]bool{certID3: true},
		}
		wantTLSCerts := []*tls.Certificate{
			&tls.Certificate{Leaf: expectCert1, Certificate: [][]byte{expectCert1.Raw}, PrivateKey: privateKey},
		}
		want509Certs := []*x509.Certificate{expectCert1, expectCert1, expectCert1}
		gotTLSCerts, got509Certs := cm.Certificates()
		if !cmp.Equal(wantTLSCerts, gotTLSCerts, cmpOpts...) {
			t.Errorf("TLS Certificates: (-want +got):\n%s", cmp.Diff(wantTLSCerts, gotTLSCerts, cmpOpts...))
		}
		if !cmp.Equal(want509Certs, got509Certs, cmpOpts...) {
			t.Errorf("x509 Certificates: (-want +got):\n%s", cmp.Diff(want509Certs, got509Certs, cmpOpts...))
		}
	})
}
