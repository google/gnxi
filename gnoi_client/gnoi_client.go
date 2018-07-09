package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type CertClient struct {
	client pb.CertificateManagementClient
}

func NewCertificateManagementClient(cc *grpc.ClientConn) *CertClient {
	return &CertClient{client: pb.NewCertificateManagementClient(cc)}
}

func (c *CertClient) Rotate(ctx context.Context) error {
	stream, err := c.client.Rotate(ctx)
	if err != nil {
		return fmt.Errorf("failed stream: %v", err)
	}

	if err := stream.Send(&pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{
			GenerateCsr: &pb.GenerateCSRRequest{
				CsrParams: &pb.CSRParams{
					Type:    pb.CertificateType_CT_X509,
					KeyType: pb.KeyType_KT_RSA,
					// MinKeySize uint32
					// CommonName           string
					// Country              string
					// State                string
					// City                 string
					// Organization         string
					// OrganizationalUnit   string
					// IpAddress            string
					// EmailId              string
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send GenerateCSRRequest: %v", err)
	}

	var req *pb.RotateCertificateResponse
	if req, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateResponse: %v", err)
	}

	genCSR := req.GetGeneratedCsr()
	if genCSR == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}
	// sign genCSR.Csr with CA

	if err := stream.Send(&pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateRequest{
				Certificate: &pb.Certificate{
					Type:        pb.CertificateType_CT_X509,
					Certificate: make([]byte, 5),
				},
				CertificateId: "blah",
				CaCertificate: []*pb.Certificate{},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}

	if req, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateResponse: %v", err)
	}
	loadCertificateResponse := req.GetLoadCertificate()
	if loadCertificateResponse == nil {
		return fmt.Errorf("expected LoadCertificateResponse, got something else")
	}

	// Verify here.

	if err := stream.Send(&pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_FinalizeRotation{FinalizeRotation: &pb.FinalizeRequest{}},
	}); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}

	return nil
}

func (c *CertClient) Install(ctx context.Context) error {
	stream, err := c.client.Install(ctx)
	if err != nil {
		return fmt.Errorf("failed stream: %v", err)
	}

	if err := stream.Send(&pb.InstallCertificateRequest{
		InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{
			GenerateCsr: &pb.GenerateCSRRequest{
				CsrParams: &pb.CSRParams{
					Type:    pb.CertificateType_CT_X509,
					KeyType: pb.KeyType_KT_RSA,
					// MinKeySize uint32
					// CommonName           string
					// Country              string
					// State                string
					// City                 string
					// Organization         string
					// OrganizationalUnit   string
					// IpAddress            string
					// EmailId              string
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send GenerateCSRRequest: %v", err)
	}

	var req *pb.InstallCertificateResponse
	if req, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateResponse: %v", err)
	}

	genCSR := req.GetGeneratedCsr()
	if genCSR == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}
	// sign genCSR.Csr with CA

	if err := stream.Send(&pb.InstallCertificateRequest{
		InstallRequest: &pb.InstallCertificateRequest_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateRequest{
				Certificate: &pb.Certificate{
					Type:        pb.CertificateType_CT_X509,
					Certificate: make([]byte, 5),
				},
				CertificateId: "blah",
				CaCertificate: []*pb.Certificate{},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}

	if req, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateResponse: %v", err)
	}
	loadCertificateResponse := req.GetLoadCertificate()
	if loadCertificateResponse == nil {
		return fmt.Errorf("expected LoadCertificateResponse, got something else")
	}

	return nil
}

func (c *CertClient) GetCertificates(ctx context.Context) error {
	_, err := c.client.GetCertificates(ctx, &pb.GetCertificatesRequest{})
	return err
}

func (c *CertClient) RevokeCertificates(ctx context.Context) error {
	_, err := c.client.RevokeCertificates(ctx, &pb.RevokeCertificatesRequest{CertificateId: []string{"id1", "id2"}})
	return err
}

func (c *CertClient) CanGenerateCSR(ctx context.Context) (bool, error) {
	result, err := c.client.CanGenerateCSR(ctx, &pb.CanGenerateCSRRequest{
		KeyType:         pb.KeyType_KT_RSA,
		CertificateType: pb.CertificateType_CT_X509,
		KeySize:         2048,
	})
	return result.CanGenerate, err
}
