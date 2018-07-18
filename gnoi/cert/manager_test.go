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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	cmpOpts = []cmp.Option{
		cmpopts.IgnoreUnexported(sync.RWMutex{}),
		cmp.AllowUnexported(x509.CertPool{}),
		cmp.AllowUnexported(Manager{}),
		cmpopts.IgnoreUnexported(Info{}),
		cmpopts.IgnoreUnexported(x509.Certificate{}),
	}
	now = time.Now()
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		wantMgr *Manager
		privKey interface{}
	}{
		{
			wantMgr: &Manager{
				privateKey: "bubu",
				certInfo:   map[string]*Info{},
				caBundle:   []*x509.Certificate{},
				locks:      map[string]bool{},
				notifiers:  []Notifier{},
			},
			privKey: "bubu",
		},
	}
	for _, test := range tests {
		gotMgr := NewManager(test.privKey)
		if !cmp.Equal(test.wantMgr, gotMgr, cmpOpts...) {
			t.Errorf("NewManager: (-want +got):\n%s", cmp.Diff(test.wantMgr, gotMgr, cmpOpts...))
		}
	}
}

func TestTLSCertificates(t *testing.T) {
	x509Cert := x509.Certificate{
		Raw: []byte{},
	}
	tlsCert := tls.Certificate{
		Leaf:        &x509Cert,
		Certificate: [][]byte{[]byte{}},
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(&x509Cert)
	ci := &Info{
		cert: &x509Cert,
	}

	tests := []struct {
		mgr          *Manager
		wantTLS      []tls.Certificate
		wantCertPool *x509.CertPool
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": ci,
					"id2": ci,
				},
				caBundle: []*x509.Certificate{&x509Cert},
			},
			wantTLS:      []tls.Certificate{tlsCert, tlsCert},
			wantCertPool: certPool,
		},
	}
	for _, test := range tests {
		gotTLS, gotPool := test.mgr.TLSCertificates()
		if !cmp.Equal(test.wantTLS, gotTLS, cmpOpts...) {
			t.Errorf("TLSCertificates: (-want +got):\n%s", cmp.Diff(test.wantTLS, gotTLS, cmpOpts...))
		}
		if !cmp.Equal(test.wantCertPool, gotPool, cmpOpts...) {
			t.Errorf("TLSCertificates: (-want +got):\n%s", cmp.Diff(test.wantCertPool, gotPool, cmpOpts...))
		}
	}
}

func TestNotifier(t *testing.T) {
	mgr := &Manager{}
	called := false
	mgr.RegisterNotifier(func(a, b int) { called = true })
	mgr.notify()

	if !called {
		t.Error("Notifyer was not called.")
	}
}

func TestRotate(t *testing.T) {
	originalDecoder := certPEMDecoder
	defer func() { certPEMDecoder = originalDecoder }()
	certPEMDecoder = func([]byte) (*x509.Certificate, error) { return &x509.Certificate{}, nil }

	tests := []struct {
		mgr        *Manager
		wantMgr    *Manager
		certID     string
		pemCert    []byte
		pemCACerts [][]byte
		wantError  bool
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id3",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id1",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  false,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id2",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
	}
	for _, test := range tests {
		_, rollback, err := test.mgr.Rotate(test.certID, test.pemCert, test.pemCACerts)
		if rollback != nil {
			rollback()
		}
		if err != nil && !test.wantError {
			t.Errorf("Rotate error: %s", err)
		}
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Rotate: (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
	}
}

func TestInstall(t *testing.T) {
	originalDecoder := certPEMDecoder
	defer func() { certPEMDecoder = originalDecoder }()
	certPEMDecoder = func([]byte) (*x509.Certificate, error) { return &x509.Certificate{}, nil }

	tests := []struct {
		mgr        *Manager
		wantMgr    *Manager
		certID     string
		pemCert    []byte
		pemCACerts [][]byte
		wantError  bool
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
					"id3": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}, &x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			certID:     "id3",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  false,
		},
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id4": &Info{},
					"id5": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id5": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id4": &Info{},
					"id5": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id5": true},
			},
			certID:     "id4",
			pemCert:    []byte{},
			pemCACerts: [][]byte{[]byte{}, []byte{}},
			wantError:  true,
		},
	}
	for _, test := range tests {
		err := test.mgr.Install(test.certID, test.pemCert, test.pemCACerts)
		if err != nil && !test.wantError {
			t.Errorf("Install error: %s", err)
		}
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Install: (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
	}
}

func TestGenCSR(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	tests := []struct {
		mgr     *Manager
		wantErr bool
	}{
		{
			mgr:     &Manager{},
			wantErr: true,
		},
		{
			mgr:     &Manager{privateKey: privKey},
			wantErr: false,
		},
	}
	for _, test := range tests {
		_, err := test.mgr.GenCSR(pkix.Name{})
		if err != nil && !test.wantErr {
			t.Errorf("GenCSR error: %s", err)
		}
	}
}

func TestGetCertInfo(t *testing.T) {
	tests := []struct {
		mgr          *Manager
		wantCertInfo []*Info
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantCertInfo: []*Info{
				&Info{},
				&Info{},
			},
		},
	}

	for _, test := range tests {
		gotCertInfo, err := test.mgr.GetCertInfo()
		if err != nil {
			t.Errorf("GetCertInfo error: %s", err)
		}
		if !cmp.Equal(test.wantCertInfo, gotCertInfo, cmpOpts...) {
			t.Errorf("GetCertInfo: (-want +got):\n%s", cmp.Diff(test.wantCertInfo, gotCertInfo, cmpOpts...))
		}
	}
}

func TestRevoke(t *testing.T) {
	tests := []struct {
		mgr     *Manager
		wantMgr *Manager
		in      []string
		wantRvk []string
		wantErr map[string]string
	}{
		{
			mgr: &Manager{
				certInfo: map[string]*Info{
					"id1": &Info{},
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			wantMgr: &Manager{
				certInfo: map[string]*Info{
					"id2": &Info{},
				},
				caBundle: []*x509.Certificate{&x509.Certificate{}},
				locks:    map[string]bool{"id2": true},
			},
			in:      []string{"id1", "id2", "id3"},
			wantRvk: []string{"id1"},
			wantErr: map[string]string{
				"id2": "an operation with this certID is in progress",
				"id3": "does not exist",
			},
		},
	}

	for _, test := range tests {
		gotRvk, gotErr, _ := test.mgr.Revoke(test.in)
		if !cmp.Equal(test.wantMgr, test.mgr, cmpOpts...) {
			t.Errorf("Revoke: Manager (-want +got):\n%s", cmp.Diff(test.wantMgr, test.mgr, cmpOpts...))
		}
		if !cmp.Equal(test.wantRvk, gotRvk) {
			t.Errorf("Revoke: revoked (-want +got):\n%s", cmp.Diff(test.wantRvk, gotRvk))
		}
		if !cmp.Equal(test.wantErr, gotErr) {
			t.Errorf("Revoke: Manager (-want +got):\n%s", cmp.Diff(test.wantErr, gotErr))
		}
	}
}
