package cert

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/gnxi/gnoi/cert/pb"

	log "github.com/golang/glog"
)

type Notifier func(int, int)

// Manager manages Certificates and CA Bundles.
type Manager struct {
	privateKey crypto.PrivateKey

	certs        map[string]*x509.Certificate
	certsModTime map[string]time.Time
	caBundle     []*x509.Certificate
	locks        map[string]bool
	notifiers    []Notifier
	mu           sync.RWMutex
}

// NewManager returns a Manager.
func NewManager(privateKey crypto.PrivateKey) *Manager {
	return &Manager{
		privateKey:   privateKey,
		certs:        map[string]*x509.Certificate{},
		certsModTime: map[string]time.Time{},
		caBundle:     []*x509.Certificate{},
		locks:        map[string]bool{},
		notifiers:    []Notifier{},
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
func (cm *Manager) Certificates() ([]tls.Certificate, *x509.CertPool) {
	certs := []tls.Certificate{}
	certPool := x509.NewCertPool()

	cm.mu.RLock()
	defer cm.mu.RUnlock()

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

// RegisterNotifier registers a function that will be called everytime the number
// of Certificates or CA certificates changes.
func (cm *Manager) RegisterNotifier(f Notifier) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.notifiers = append(cm.notifiers, f)
}

func (cm *Manager) notify() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	log.Infof("Notifying for: %d Certificates and %d CA Certificates.", len(cm.certs), len(cm.caBundle))
	for _, notifier := range cm.notifiers {
		notifier(len(cm.certs), len(cm.caBundle))
	}
}

// Empty returns true if there are no certificates, no ca certificates installed and
// no changes in progress.
func (cm *Manager) Empty() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.certs) == 0 && len(cm.caBundle) == 0 && len(cm.locks) == 0
}

var createCSR = func(rand io.Reader, template *x509.CertificateRequest, priv interface{}) ([]byte, error) {
	der, err := x509.CreateCertificateRequest(rand, template, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: der,
	}), nil
}

// GenCSR generates and returns a CSR based on the provided parameters.
func (cm *Manager) GenCSR(params *pb.CSRParams) (*pb.CSR, error) {
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
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	pemCSR, err := createCSR(rand.Reader, template, cm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}
	return &pb.CSR{
		Type: pb.CertificateType_CT_X509,
		Csr:  pemCSR,
	}, nil
}

// Get returns all the Certificates and their Certificate IDs.
func (cm *Manager) Get() ([]*pb.CertificateInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

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
func (cm *Manager) Install(r *pb.LoadCertificateRequest) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	certID := r.CertificateId

	if _, locked := cm.locks[certID]; locked {
		err := fmt.Errorf("an operation with certID %q is already in progress", certID)
		log.Error(err)
		return err
	}

	certificate, err := DecodeCert(r.Certificate.Certificate)
	if err != nil {
		return fmt.Errorf("failed to decode Certificate: %v", err)
	}

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

	go cm.notify()
	return nil
}

// Revoke revokes Certificates.
func (cm *Manager) Revoke(req *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	resp := &pb.RevokeCertificatesResponse{
		RevokedCertificateId:       []string{},
		CertificateRevocationError: []*pb.CertificateRevocationError{},
	}

	for _, certID := range req.CertificateId {
		if _, locked := cm.locks[certID]; locked {
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

	go cm.notify()
	return resp, nil
}

// Rotate rotates Certificates and CA Bundles.
func (cm *Manager) Rotate(r *pb.LoadCertificateRequest) (func(), func(), error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	certID := r.CertificateId

	if _, locked := cm.locks[certID]; locked {
		err := fmt.Errorf("an operation with certID %q is already in progress", certID)
		log.Error(err)
		return nil, nil, err
	}
	cm.locks[certID] = true

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
		delete(cm.locks, certID)
		go cm.notify()
	}

	accept := func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		delete(cm.locks, certID)
	}

	go cm.notify()
	return accept, rollback, nil
}
