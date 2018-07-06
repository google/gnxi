package gnoi

import (
	"golang.org/x/net/context"

	pb "github.com/google/gnxi/gnoi/proto"
)

type server struct{}

func (s *server) Rotate(stream pb.CertificateManagement_RotateServer) error {
	return nil
}
func (s *server) Install(stream pb.CertificateManagement_InstallServer) error {
	return nil
}

func (s *server) GetCertificates(ctx context.Context, helloRequest *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	return &pb.GetCertificatesResponse{}, nil
}

func (s *server) RevokeCertificates(ctx context.Context, helloRequest *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	return &pb.RevokeCertificatesResponse{}, nil
}

func (s *server) CanGenerateCSR(ctx context.Context, helloRequest *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	return &pb.CanGenerateCSRResponse{}, nil
}
