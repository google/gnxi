package cert

import (
	"fmt"

	"github.com/google/gnxi/gnoi/cert/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/golang/glog"
)

// CertInterface provides the necessary methods to handle the Certificate Management service.
type ManagerInterface interface {
	GenCSR(*pb.CSRParams) (*pb.CSR, error)
	Get() ([]*pb.CertificateInfo, error)
	Install(*pb.LoadCertificateRequest) error
	Revoke(*pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error)
	Rotate(*pb.LoadCertificateRequest) (func(), func(), error)
}

// Server is a Certificate Management service.
type Server struct {
	manager ManagerInterface
}

// NewServer returns a Certificate Management Server.
func NewServer(manager ManagerInterface) *Server {
	return &Server{manager: manager}
}

// Register registers the server into the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterCertificateManagementServer(g, s)
}

// Rotate allows rotating a certificate.
func (s *Server) Rotate(stream pb.CertificateManagement_RotateServer) error {
	var resp *pb.RotateCertificateRequest
	var err error

	log.Info("Start Rotate request.")

	if resp, err = stream.Recv(); err != nil {
		rerr := fmt.Errorf("failed to receive RotateCertificateRequest: %v", err)
		log.Error(rerr)
		return rerr
	}
	genCSRRequest := resp.GetGenerateCsr()
	if genCSRRequest == nil {
		rerr := fmt.Errorf("expected GenerateCSRRequest, got something else")
		log.Error(rerr)
		return rerr
	}

	csr, err := s.manager.GenCSR(genCSRRequest.CsrParams)
	if err != nil {
		rerr := fmt.Errorf("failed to generate CSR: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err = stream.Send(&pb.RotateCertificateResponse{
		RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: csr},
		},
	}); err != nil {
		rerr := fmt.Errorf("failed to send GenerateCSRResponse: %v", err)
		log.Error(rerr)
		return rerr
	}

	if resp, err = stream.Recv(); err != nil {
		rerr := fmt.Errorf("failed to receive RotateCertificateRequest: %v", err)
		log.Error(rerr)
		return rerr
	}
	loadCertificateRequest := resp.GetLoadCertificate()
	if loadCertificateRequest == nil {
		rerr := fmt.Errorf("expected LoadCertificateRequest, got something else")
		log.Error(rerr)
		return rerr
	}

	rotateAccept, rotateBack, err := s.manager.Rotate(loadCertificateRequest)
	if err != nil {
		rerr := fmt.Errorf("failed to load the Certificate: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err = stream.Send(&pb.RotateCertificateResponse{
		RotateResponse: &pb.RotateCertificateResponse_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateResponse{},
		},
	}); err != nil {
		rerr := fmt.Errorf("failed to send LoadCertificateResponse: %v", err)
		log.Error(rerr)
		return rerr
	}

	if resp, err = stream.Recv(); err != nil {
		rotateBack()
		rerr := fmt.Errorf("rolling back - failed to receive RotateCertificateRequest: %v", err)
		log.Error(rerr)
		return rerr
	}
	finalize := resp.GetFinalizeRotation()
	if finalize == nil {
		rotateBack()
		rerr := fmt.Errorf("expected FinalizeRequest, got something else")
		log.Error(rerr)
		return rerr
	}
	rotateAccept()
	log.Info("Success Rotate request.")

	return nil
}

// Install installs a certificate.
func (s *Server) Install(stream pb.CertificateManagement_InstallServer) error {
	var resp *pb.InstallCertificateRequest
	var err error
	log.Info("Start Install request.")

	if resp, err = stream.Recv(); err != nil {
		rerr := fmt.Errorf("failed to receive InstallCertificateRequest: %v", err)
		log.Error(rerr)
		return rerr
	}
	genCSRRequest := resp.GetGenerateCsr()
	if genCSRRequest == nil {
		rerr := fmt.Errorf("expected GenerateCSRRequest, got something else")
		log.Error(rerr)
		return rerr
	}

	csr, err := s.manager.GenCSR(genCSRRequest.CsrParams)
	if err != nil {
		rerr := fmt.Errorf("failed to generate CSR: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err = stream.Send(&pb.InstallCertificateResponse{
		InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: csr},
		},
	}); err != nil {
		rerr := fmt.Errorf("failed to send GenerateCSRResponse: %v", err)
		log.Error(rerr)
		return rerr
	}

	if resp, err = stream.Recv(); err != nil {
		rerr := fmt.Errorf("failed to receive InstallCertificateRequest: %v", err)
		log.Error(rerr)
		return rerr
	}
	loadCertificateRequest := resp.GetLoadCertificate()
	if loadCertificateRequest == nil {
		rerr := fmt.Errorf("expected LoadCertificateRequest, got something else")
		log.Error(rerr)
		return rerr
	}
	if err := s.manager.Install(loadCertificateRequest); err != nil {
		rerr := fmt.Errorf("failed to load the Certificate: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err := stream.Send(&pb.InstallCertificateResponse{
		InstallResponse: &pb.InstallCertificateResponse_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateResponse{},
		},
	}); err != nil {
		rerr := fmt.Errorf("failed to send LoadCertificateResponse: %v", err)
		log.Error(rerr)
		return rerr
	}

	log.Info("Success Install request.")
	return nil
}

// GetCertificates returns installed certificates.
func (s *Server) GetCertificates(ctx context.Context, request *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	log.Info("GetCertificates request.")
	certInfo, err := s.manager.Get()
	return &pb.GetCertificatesResponse{CertificateInfo: certInfo}, err
}

// RevokeCertificates revokes certificates.
func (s *Server) RevokeCertificates(ctx context.Context, request *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	log.Info("RevokeCertificates request.")
	return s.manager.Revoke(request)
}

// CanGenerateCSR returns if it can generate CSRs with the given properties.
func (s *Server) CanGenerateCSR(ctx context.Context, request *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	log.Info("CanGenerateCSR request.")
	return &pb.CanGenerateCSRResponse{CanGenerate: request.KeyType == pb.KeyType_KT_RSA && request.CertificateType == pb.CertificateType_CT_X509 && request.KeySize >= 128}, nil
}
