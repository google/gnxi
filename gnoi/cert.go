package gnoi

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/google/gnxi/gnoi/certpb"
)

// CertInterface provides the necessary methods to handle the Certificate Management service.
type CertInterface interface {
	GenCSR(*pb.CSRParams) (*pb.CSR, error)
	Get() ([]*pb.CertificateInfo, error)
	Install(*pb.LoadCertificateRequest) error
	Revoke(*pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error)
	Rotate(*pb.LoadCertificateRequest) (func(), func(), error)
}

// CertServer is a CertificateManagement service.
type CertServer struct {
	certInterface CertInterface
}

// NewCertServer returns a CertServer.
func NewCertServer(certiInterface CertInterface) *CertServer {
	return &CertServer{certInterface: certiInterface}
}

// Register registers the server into the gRPC server provided.
func (s *CertServer) Register(g *grpc.Server) {
	pb.RegisterCertificateManagementServer(g, s)
}

// Rotate allows rotating a certificate.
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

// Install installs a certificate.
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

// GetCertificates returns installed certificates.
func (s *CertServer) GetCertificates(ctx context.Context, request *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	certInfo, err := s.certInterface.Get()
	return &pb.GetCertificatesResponse{CertificateInfo: certInfo}, err
}

// RevokeCertificates revokes certificates.
func (s *CertServer) RevokeCertificates(ctx context.Context, request *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	return s.certInterface.Revoke(request)
}

// CanGenerateCSR returns if it can generate CSRs with the given properties.
func (s *CertServer) CanGenerateCSR(ctx context.Context, request *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	return &pb.CanGenerateCSRResponse{CanGenerate: request.KeyType == pb.KeyType_KT_RSA && request.CertificateType == pb.CertificateType_CT_X509 && request.KeySize >= 128}, nil
}

// CertManager manages Certificates and CA Bundles.
type CertManager struct {
	privateKey crypto.PrivateKey

	certs        map[string]*x509.Certificate
	certsModTime map[string]time.Time
	caBundle     []*x509.Certificate
	mu           sync.Mutex

	locks  map[string]bool
	muLock sync.Mutex
}

// NewCertManager returns a CertManager.
func NewCertManager(privateKey crypto.PrivateKey) *CertManager {
	return &CertManager{
		privateKey:   privateKey,
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

// Certificates returns the list of Certificates and CA certificates.
func (cm *CertManager) Certificates() ([]tls.Certificate, *x509.CertPool) {
	certs := []tls.Certificate{}
	certPool := x509.NewCertPool()

	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, cert := range cm.certs {
		certs = append(certs, tls.Certificate{
			Leaf:        cert,
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  cm.privateKey,
		})
	}
	for _, cert := range cm.caBundle {
		certPool.AddCert(cert)
	}
	return certs, certPool
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

var createCSR = x509.CreateCertificateRequest

// GenCSR generates and returns a CSR based on the provided parameters.
func (cm *CertManager) GenCSR(params *pb.CSRParams) (*pb.CSR, error) {
	if params.Type != pb.CertificateType_CT_X509 {
		return nil, fmt.Errorf("certificate type %q not supported", params.Type)
	}
	if params.KeyType != pb.KeyType_KT_RSA {
		return nil, fmt.Errorf("key type %q not supported", params.KeyType)
	}

	subject := pkix.Name{
		Country:            []string{params.Country},
		Organization:       []string{params.Organization},
		OrganizationalUnit: []string{params.OrganizationalUnit},
		CommonName:         params.CommonName,
	}
	template := &x509.CertificateRequest{
		Subject:            subject,
		EmailAddresses:     []string{params.EmailId},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	address := net.ParseIP(params.IpAddress)
	if address != nil {
		template.IPAddresses = []net.IP{address}
	} else {
		template.DNSNames = []string{params.IpAddress}
	}
	csr, err := createCSR(rand.Reader, template, cm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}
	return &pb.CSR{
		Type: pb.CertificateType_CT_X509,
		Csr:  csr,
	}, nil
}

// Get returns all the Certificates and their Certificate IDs.
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

// Install installs new Certificates and CA Bundles.
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

// Revoke revokes Certificates.
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

// Rotate rotates Certificates and CA Bundles.
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

// CertClient is a Certificate Management service client.
type CertClient struct {
	client pb.CertificateManagementClient
}

// NewCertClient returns a new CertClient.
func NewCertClient(c *grpc.ClientConn) *CertClient {
	return &CertClient{client: pb.NewCertificateManagementClient(c)}
}

// Rotate rotates a certificate.
func (c *CertClient) Rotate(ctx context.Context, params pkix.Name, sign func(*x509.CertificateRequest) *x509.Certificate, validate func() bool) error {
	return nil
}

// Install installs a certificate.
func (c *CertClient) Install(ctx context.Context, params pkix.Name, signer func(*x509.CertificateRequest) *x509.Certificate) error {
	return nil
}

// GetCertificates gets a map of certificates in the target, certID to certificate
func (c *CertClient) GetCertificates(ctx context.Context) (map[string]x509.Certificate, error) {
	return nil, nil
}

// RevokeCertificates revokes certificates in the target, returns a map of certID to error for the ones that failed to be revoked.
func (c *CertClient) RevokeCertificates(ctx context.Context, certIDs []string) (map[string]string, error) {
	return nil, nil
}

// CanGenerateCSR checks if the target can generate a CSR.
func (c *CertClient) CanGenerateCSR(ctx context.Context) (bool, error) {
	return false, nil
}
