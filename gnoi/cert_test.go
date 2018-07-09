package gnoi

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/google/gnxi/gnoi/certpb"
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
)

type mockCertInterface struct {
	CertInterface
}

func (m *mockCertInterface) Lock(cID string) error {
	return nil
}

func (m *mockCertInterface) Unlock(cID string) {}

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

func (m *mockCertInterface) Rotate(l *pb.LoadCertificateRequest) (func(), error) {
	if !Equal(expectLoadCertificateRequest, l) {
		return nil, fmt.Errorf("Rotate: (-want +got):\n%s", Diff(expectLoadCertificateRequest, l))
	}
	return func() {}, nil
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
		if err := stream.Send(&pb.RotateCertificateRequest{
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
		if err := stream.Send(&pb.RotateCertificateRequest{
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

		if err := stream.Send(&pb.InstallCertificateRequest{
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

		if err := stream.Send(&pb.InstallCertificateRequest{
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
	cm := NewCertManager()

	t.Run("Lock & Unlock", func(t *testing.T) {
		certID := "test"
		if err := cm.Lock(certID); err != nil {
			t.Errorf("failed to lock: %v", err)
		}
		if err := cm.Lock(certID); err == nil {
			t.Errorf("allowed double lock")
		}
		cm.Unlock(certID)
		if err := cm.Lock(certID); err != nil {
			t.Errorf("failed to lock: %v", err)
		}
	})

}
