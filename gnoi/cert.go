package gnoi

import (
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"

	pb "github.com/google/gnxi/gnoi/certpb"
)

type CertInterface interface {
	Lock(string) error
	Unlock(string)
	GenCSR(*pb.CSRParams) (*pb.CSR, error)
	Get() ([]*pb.CertificateInfo, error)
	Install(*pb.LoadCertificateRequest) error
	Revoke(*pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error)
	Rotate(*pb.LoadCertificateRequest) (func(), error)
}

type CertServer struct {
	certInterface CertInterface
}

func NewCertServer(cm CertInterface) *CertServer {
	return &CertServer{certInterface: cm}
}

func (s *CertServer) Rotate(stream pb.CertificateManagement_RotateServer) error {
	var resp *pb.RotateCertificateRequest
	var err error

	if resp, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateRequest: %v", err)
	}
	genCSRRequest := resp.GetGenerateCsr()
	if genCSRRequest == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}

	if err := s.certInterface.Lock(genCSRRequest.CertificateId); err != nil {
		return fmt.Errorf("there is already an ongoing operation with this Certificate ID: %v", err)
	}
	defer s.certInterface.Unlock(genCSRRequest.CertificateId)

	csr, err := s.certInterface.GenCSR(genCSRRequest.CsrParams)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %v", err)
	}

	if err = stream.Send(&pb.RotateCertificateResponse{
		RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: csr},
		},
	}); err != nil {
		return fmt.Errorf("failed to send GenerateCSRResponse: %v", err)
	}

	if resp, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateRequest: %v", err)
	}
	loadCertificateRequest := resp.GetLoadCertificate()
	if loadCertificateRequest == nil {
		return fmt.Errorf("expected LoadCertificateRequest, got something else")
	}

	rotateBack, err := s.certInterface.Rotate(loadCertificateRequest)
	if err != nil {
		return fmt.Errorf("failed to load the Certificate: %v", err)
	}

	if err := stream.Send(&pb.RotateCertificateResponse{
		RotateResponse: &pb.RotateCertificateResponse_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateResponse{},
		},
	}); err != nil {
		return fmt.Errorf("failed to send LoadCertificateResponse: %v", err)
	}

	if resp, err = stream.Recv(); err != nil {
		rotateBack()
		return fmt.Errorf("rolling back - failed to receive RotateCertificateRequest: %v", err)
	}
	finalize := resp.GetFinalizeRotation()
	if finalize == nil {
		return fmt.Errorf("expected FinalizeRequest, got something else")
	}

	return nil
}

func (s *CertServer) Install(stream pb.CertificateManagement_InstallServer) error {
	var resp *pb.InstallCertificateRequest
	var err error

	if resp, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateRequest: %v", err)
	}
	genCSRRequest := resp.GetGenerateCsr()
	if genCSRRequest == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}

	if err := s.certInterface.Lock(genCSRRequest.CertificateId); err != nil {
		return fmt.Errorf("there is already an ongoing operation with this Certificate ID: %v", err)
	}
	defer s.certInterface.Unlock(genCSRRequest.CertificateId)

	csr, err := s.certInterface.GenCSR(genCSRRequest.CsrParams)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %v", err)
	}

	if err = stream.Send(&pb.InstallCertificateResponse{
		InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: csr},
		},
	}); err != nil {
		return fmt.Errorf("failed to send GenerateCSRResponse: %v", err)
	}

	if resp, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateRequest: %v", err)
	}
	loadCertificateRequest := resp.GetLoadCertificate()
	if loadCertificateRequest == nil {
		return fmt.Errorf("expected LoadCertificateRequest, got something else")
	}
	if err := s.certInterface.Install(loadCertificateRequest); err != nil {
		return fmt.Errorf("failed to load the Certificate: %v", err)
	}

	if err := stream.Send(&pb.InstallCertificateResponse{
		InstallResponse: &pb.InstallCertificateResponse_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateResponse{},
		},
	}); err != nil {
		return fmt.Errorf("failed to send LoadCertificateResponse: %v", err)
	}
	return nil
}

func (s *CertServer) GetCertificates(ctx context.Context, request *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	certInfo, err := s.certInterface.Get()
	return &pb.GetCertificatesResponse{CertificateInfo: certInfo}, err
}

func (s *CertServer) RevokeCertificates(ctx context.Context, request *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	return s.certInterface.Revoke(request)
}

func (s *CertServer) CanGenerateCSR(ctx context.Context, request *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	return &pb.CanGenerateCSRResponse{CanGenerate: request.KeyType == pb.KeyType_KT_RSA && request.CertificateType == pb.CertificateType_CT_X509 && request.KeySize >= 128}, nil
}

type CertManager struct {
	certs        map[string]*x509.Certificate
	certsModTime map[string]time.Time
	muCerts      sync.Mutex

	caBundle   []*x509.Certificate
	muCABundle sync.Mutex

	lock   map[string]bool
	muLock sync.Mutex
}

func NewCertManager() *CertManager {
	return &CertManager{
		certs:        map[string]*x509.Certificate{},
		certsModTime: map[string]time.Time{},
		caBundle:     []*x509.Certificate{},
		lock:         map[string]bool{},
	}
}

func (cm *CertManager) Lock(certID string) error {
	cm.muLock.Lock()
	defer cm.muLock.Unlock()
	if _, ok := cm.lock[certID]; ok {
		return fmt.Errorf("an operation with certID %q is already in progress", certID)
	}
	cm.lock[certID] = true
	return nil
}

func (cm *CertManager) Unlock(certID string) {
	cm.muLock.Lock()
	defer cm.muLock.Unlock()
	delete(cm.lock, certID)
	return
}

func (cm *CertManager) GenCSR(params *pb.CSRParams) (*pb.CSR, error) {
	return nil, nil
}

func (cm *CertManager) Get() ([]*pb.CertificateInfo, error) {
	cm.muCerts.Lock()
	defer cm.muCerts.Unlock()
	r := []*pb.CertificateInfo{}
	for certID, cert := range cm.certs {
		r = append(r, &pb.CertificateInfo{
			CertificateId: certID,
			Certificate: &pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: cert.Raw, // needs encoding !!!!!!!!!!!!!!!!
			},
			ModificationTime: cm.certsModTime[certID].UnixNano(),
		})
	}
	return r, nil
}

func (cm *CertManager) Install(r *pb.LoadCertificateRequest) error {
	return nil
}
func (cm *CertManager) Revoke(req *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	cm.muCerts.Lock()
	defer cm.muCerts.Unlock()
	resp := &pb.RevokeCertificatesResponse{
		RevokedCertificateId:       []string{},
		CertificateRevocationError: []*pb.CertificateRevocationError{},
	}
	for _, certID := range req.CertificateId {
		if _, ok := cm.certs[certID]; ok {
			delete(cm.certs, certID)
			delete(cm.certsModTime, certID)
			resp.RevokedCertificateId = append(resp.RevokedCertificateId, certID)
			continue
		}
		// add to failed
	}
	return nil, nil
}
func (cm *CertManager) Rotate(r *pb.LoadCertificateRequest) (func(), error) {
	return func() {}, nil
}
