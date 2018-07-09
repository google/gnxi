package gnoi

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"

	pb "github.com/google/gnxi/gnoi/certpb"
)

// CertInterface does blah.
type CertInterface interface {
	GenCSR(*pb.CSRParams) (*pb.CSR, error)
	Get() ([]*pb.CertificateInfo, error)
	Install(*pb.LoadCertificateRequest) error
	Revoke(*pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error)
	Rotate(*pb.LoadCertificateRequest) (func(), func(), error)
}

// CertServer does blah.
type CertServer struct {
	certInterface CertInterface
}

// NewCertServer does blah.
func NewCertServer(cm CertInterface) *CertServer {
	return &CertServer{certInterface: cm}
}

// Rotate does blah.
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

	rotateAccept, rotateBack, err := s.certInterface.Rotate(loadCertificateRequest)
	if err != nil {
		return fmt.Errorf("failed to load the Certificate: %v", err)
	}

	if err = stream.Send(&pb.RotateCertificateResponse{
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
		rotateBack()
		return fmt.Errorf("expected FinalizeRequest, got something else")
	}
	rotateAccept()

	return nil
}

// Install does blah.
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

// GetCertificates does blah.
func (s *CertServer) GetCertificates(ctx context.Context, request *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	certInfo, err := s.certInterface.Get()
	return &pb.GetCertificatesResponse{CertificateInfo: certInfo}, err
}

// RevokeCertificates does blah.
func (s *CertServer) RevokeCertificates(ctx context.Context, request *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	return s.certInterface.Revoke(request)
}

// CanGenerateCSR does blah.
func (s *CertServer) CanGenerateCSR(ctx context.Context, request *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	return &pb.CanGenerateCSRResponse{CanGenerate: request.KeyType == pb.KeyType_KT_RSA && request.CertificateType == pb.CertificateType_CT_X509 && request.KeySize >= 128}, nil
}

// CertManager does blah.
type CertManager struct {
	privateKey crypto.PrivateKey

	certs        map[string]*x509.Certificate
	certsModTime map[string]time.Time
	caBundle     []*x509.Certificate
	mu           sync.Mutex

	locks  map[string]bool
	muLock sync.Mutex
}

// NewCertManager does blah.
func NewCertManager(p crypto.PrivateKey) *CertManager {
	return &CertManager{
		privateKey:   p,
		certs:        map[string]*x509.Certificate{},
		certsModTime: map[string]time.Time{},
		caBundle:     []*x509.Certificate{},
		locks:        map[string]bool{},
	}
}

func toSlices(certs []*pb.Certificate) [][]byte {
	ret := [][]byte{}
	for _, cert := range certs {
		ret = append(ret, cert.Certificate)
	}
	return ret
}

var nowTime = time.Now

// Certificates does blah.
func (cm *CertManager) Certificates() ([]*tls.Certificate, []*x509.Certificate) {
	certs := []*tls.Certificate{}
	caBundle := []*x509.Certificate{}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, cert := range cm.certs {
		certs = append(certs, &tls.Certificate{
			Leaf:        cert,
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  cm.privateKey,
		})
	}
	for _, cert := range cm.caBundle {
		caBundle = append(caBundle, cert)
	}
	return certs, caBundle
}

func (cm *CertManager) locked(certID string) bool {
	cm.muLock.Lock()
	defer cm.muLock.Unlock()
	_, ok := cm.locks[certID]
	return ok
}

func (cm *CertManager) lock(certID string) error {
	cm.muLock.Lock()
	defer cm.muLock.Unlock()
	if _, ok := cm.locks[certID]; ok {
		return fmt.Errorf("an operation with certID %q is already in progress", certID)
	}
	cm.locks[certID] = true
	return nil
}

func (cm *CertManager) unlock(certID string) {
	cm.muLock.Lock()
	defer cm.muLock.Unlock()
	delete(cm.locks, certID)
	return
}

// GenCSR does blah.
func (cm *CertManager) GenCSR(params *pb.CSRParams) (*pb.CSR, error) {

	return nil, nil
}

// Get does blah.
func (cm *CertManager) Get() ([]*pb.CertificateInfo, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	r := []*pb.CertificateInfo{}
	for certID, cert := range cm.certs {
		r = append(r, &pb.CertificateInfo{
			CertificateId: certID,
			Certificate: &pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: EncodeCert(cert),
			},
			ModificationTime: cm.certsModTime[certID].UnixNano(),
		})
	}
	return r, nil
}

// Install does blah.
func (cm *CertManager) Install(r *pb.LoadCertificateRequest) error {
	certID := r.CertificateId
	if err := cm.lock(certID); err != nil {
		return err
	}
	defer cm.unlock(certID)

	certificate, err := DecodeCert(r.Certificate.Certificate)
	if err != nil {
		return fmt.Errorf("failed to decode Certificate: %v", err)
	}
	cm.mu.Lock()
	cm.mu.Unlock()
	if _, ok := cm.certs[certID]; ok {
		return fmt.Errorf("certificate id %q already exists", certID)
	}
	cm.certs[certID] = certificate
	cm.certsModTime[certID] = nowTime()

	if r.CaCertificate != nil && len(r.CaCertificate) != 0 {
		certificates, err := DecodeCerts(toSlices(r.CaCertificate))
		if err != nil {
			return fmt.Errorf("failed to decode CA Bundle: %v", err)
		}
		cm.caBundle = certificates
	}

	return nil
}

// Revoke does blah.
func (cm *CertManager) Revoke(req *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	resp := &pb.RevokeCertificatesResponse{
		RevokedCertificateId:       []string{},
		CertificateRevocationError: []*pb.CertificateRevocationError{},
	}
	for _, certID := range req.CertificateId {
		if cm.locked(certID) {
			revErr := &pb.CertificateRevocationError{
				CertificateId: certID,
				ErrorMessage:  "an operation with this certID is in progress",
			}
			resp.CertificateRevocationError = append(resp.CertificateRevocationError, revErr)
			continue
		}
		if _, ok := cm.certs[certID]; ok {
			delete(cm.certs, certID)
			delete(cm.certsModTime, certID)
			resp.RevokedCertificateId = append(resp.RevokedCertificateId, certID)
			continue
		}
		revErr := &pb.CertificateRevocationError{
			CertificateId: certID,
			ErrorMessage:  "does not exist",
		}
		resp.CertificateRevocationError = append(resp.CertificateRevocationError, revErr)
	}
	return resp, nil
}

// Rotate does blah.
func (cm *CertManager) Rotate(r *pb.LoadCertificateRequest) (func(), func(), error) {
	certID := r.CertificateId
	if err := cm.lock(certID); err != nil {
		return nil, nil, err
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	oldCert, ok := cm.certs[certID]
	if !ok {
		return nil, nil, fmt.Errorf("certificate ID %q does not exist", certID)
	}
	certificate, err := DecodeCert(r.Certificate.Certificate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Certificate: %v", err)
	}
	cm.certs[certID] = certificate
	cm.certsModTime[certID] = nowTime()

	var oldCABundle []*x509.Certificate
	if r.CaCertificate != nil && len(r.CaCertificate) != 0 {
		certificates, err := DecodeCerts(toSlices(r.CaCertificate))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode CA Bundle: %v", err)
		}
		oldCABundle = cm.caBundle
		cm.caBundle = certificates
	}

	rollback := func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		cm.certs[certID] = oldCert
		if oldCABundle != nil {
			cm.caBundle = oldCABundle
		}
		cm.unlock(certID)
	}

	accept := func() {
		cm.unlock(certID)
	}

	return accept, rollback, nil
}
