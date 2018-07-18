/* Copyright 2018 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	log "github.com/golang/glog"
)

// Info contains information about a x509 Certificate.
type Info struct {
	certID  string
	cert    *x509.Certificate
	updated time.Time
}

// Notifier is called with number of Certificates and CA Certificates.
type Notifier func(int, int)

// Manager manages Certificates and CA Bundles.
type Manager struct {
	privateKey crypto.PrivateKey

	certInfo  map[string]*Info
	caBundle  []*x509.Certificate
	locks     map[string]bool
	notifiers []Notifier
	mu        sync.RWMutex
}

// NewManager returns a Manager.
func NewManager(privateKey crypto.PrivateKey) *Manager {
	return &Manager{
		privateKey: privateKey,
		certInfo:   map[string]*Info{},
		caBundle:   []*x509.Certificate{},
		locks:      map[string]bool{},
		notifiers:  []Notifier{},
	}
}

var nowTime = time.Now

// TLSCertificates returns a list of TLS Certificates and a x509 Pool of CA Certificates.
func (cm *Manager) TLSCertificates() ([]tls.Certificate, *x509.CertPool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	certs := []tls.Certificate{}
	certPool := x509.NewCertPool()
	for _, ci := range cm.certInfo {
		certs = append(certs, tls.Certificate{
			Leaf:        ci.cert,
			Certificate: [][]byte{ci.cert.Raw},
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

	log.Infof("Notifying for: %d Certificates and %d CA Certificates.", len(cm.certInfo), len(cm.caBundle))
	for _, notifier := range cm.notifiers {
		// This is a blocking call to allow the caller to be sure the read locks are active.
		notifier(len(cm.certInfo), len(cm.caBundle))
	}
}

// PEMtox509 decodes a PEM block into a x509.Certificate.
func PEMtox509(bytes []byte) (*x509.Certificate, error) {
	certDERBlock, _ := pem.Decode(bytes)
	if certDERBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	certificate, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode DER bytes")
	}
	return certificate, nil
}

var certPEMDecoder = PEMtox509

// update installs or rotates a Certificate.
func (cm *Manager) update(requireExisting bool, certID string, pemCert []byte, pemCACerts [][]byte) (func(), func(), error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.certInfo[certID]; ok && !requireExisting {
		return nil, nil, fmt.Errorf("certificate id %q already exists", certID)
	} else if !ok && requireExisting {
		return nil, nil, fmt.Errorf("certificate ID %q does not exist", certID)
	}

	if _, locked := cm.locks[certID]; locked {
		return nil, nil, fmt.Errorf("an operation with certID %q is already in progress", certID)
	}
	cm.locks[certID] = true

	x509Cert, err := certPEMDecoder(pemCert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Certificate: %v", err)
	}

	oldCertInfo := cm.certInfo[certID]
	cm.certInfo[certID] = &Info{
		cert:    x509Cert,
		updated: nowTime(),
		certID:  certID,
	}

	var oldCABundle []*x509.Certificate
	if len(pemCACerts) != 0 {
		newBundle := []*x509.Certificate{}
		for _, pem := range pemCACerts {
			x509Cert, err := certPEMDecoder(pem)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode cert in CA Bundle: %v", err)
			}
			newBundle = append(newBundle, x509Cert)
		}
		oldCABundle = cm.caBundle
		cm.caBundle = newBundle
	}

	rollback := func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		cm.certInfo[certID] = oldCertInfo
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

// Install installs new Certificates and optionally updates the CA Bundles.
func (cm *Manager) Install(certID string, pemCert []byte, pemCACerts [][]byte) error {
	accept, _, err := cm.update(false, certID, pemCert, pemCACerts)
	if err != nil {
		return err
	}
	accept()
	return nil
}

// Rotate rotates Certificates and optionally updates the CA Bundles.
func (cm *Manager) Rotate(certID string, pemCert []byte, pemCACerts [][]byte) (func(), func(), error) {
	return cm.update(true, certID, pemCert, pemCACerts)
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
func (cm *Manager) GenCSR(subject pkix.Name) ([]byte, error) {
	template := &x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	pemCSR, err := createCSR(rand.Reader, template, cm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}
	return pemCSR, nil
}

// GetCertInfo returns all the Certificates, Certificate IDs and updated times.
func (cm *Manager) GetCertInfo() ([]*Info, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	certInfo := []*Info{}
	for _, ci := range cm.certInfo {
		certInfo = append(certInfo, ci)
	}

	return certInfo, nil
}

// Revoke revokes Certificates.
func (cm *Manager) Revoke(revoke []string) ([]string, map[string]string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	revoked := []string{}
	notRevoked := map[string]string{}

	for _, certID := range revoke {
		if _, locked := cm.locks[certID]; locked {
			notRevoked[certID] = "an operation with this certID is in progress"
			continue
		}
		if _, ok := cm.certInfo[certID]; ok {
			delete(cm.certInfo, certID)
			revoked = append(revoked, certID)
			continue
		}
		notRevoked[certID] = "does not exist"
	}

	go cm.notify()
	return revoked, notRevoked, nil
}
