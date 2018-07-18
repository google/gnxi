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
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"

	"github.com/google/gnxi/gnoi/cert/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/golang/glog"
)

// ManagerInterface provides the necessary methods to handle the Certificate Management service.
type ManagerInterface interface {
	Install(string, []byte, [][]byte) error
	Rotate(string, []byte, [][]byte) (func(), func(), error)
	GenCSR(pkix.Name) ([]byte, error)
	GetCertInfo() ([]*Info, error)
	Revoke([]string) ([]string, map[string]string, error)
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

	if genCSRRequest.CsrParams.Type != pb.CertificateType_CT_X509 {
		return fmt.Errorf("certificate type %q not supported", genCSRRequest.CsrParams.Type)
	}
	if genCSRRequest.CsrParams.KeyType != pb.KeyType_KT_RSA {
		return fmt.Errorf("key type %q not supported", genCSRRequest.CsrParams.KeyType)
	}
	subject := pkix.Name{
		Country:            []string{genCSRRequest.CsrParams.Country},
		Organization:       []string{genCSRRequest.CsrParams.Organization},
		OrganizationalUnit: []string{genCSRRequest.CsrParams.OrganizationalUnit},
		CommonName:         genCSRRequest.CsrParams.CommonName,
	}

	pemCSR, err := s.manager.GenCSR(subject)
	if err != nil {
		rerr := fmt.Errorf("failed to generate CSR: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err = stream.Send(&pb.InstallCertificateResponse{
		InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{
				Type: pb.CertificateType_CT_X509,
				Csr:  pemCSR,
			}},
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

	if loadCertificateRequest.Certificate.Type != pb.CertificateType_CT_X509 {
		rerr := fmt.Errorf("unexpected Certificate type: %d", loadCertificateRequest.Certificate.Type)
		log.Error(rerr)
		return rerr
	}

	certID := loadCertificateRequest.CertificateId
	pemCert := loadCertificateRequest.Certificate.Certificate
	pemCACerts := [][]byte{}
	for _, cert := range loadCertificateRequest.CaCertificate {
		if cert.Type != pb.CertificateType_CT_X509 {
			rerr := fmt.Errorf("unexpected Certificate type: %d", cert.Type)
			log.Error(rerr)
			return rerr
		}
		pemCACerts = append(pemCACerts, cert.Certificate)
	}

	if err := s.manager.Install(certID, pemCert, pemCACerts); err != nil {
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

	if genCSRRequest.CsrParams.Type != pb.CertificateType_CT_X509 {
		return fmt.Errorf("certificate type %q not supported", genCSRRequest.CsrParams.Type)
	}
	if genCSRRequest.CsrParams.KeyType != pb.KeyType_KT_RSA {
		return fmt.Errorf("key type %q not supported", genCSRRequest.CsrParams.KeyType)
	}
	subject := pkix.Name{
		Country:            []string{genCSRRequest.CsrParams.Country},
		Organization:       []string{genCSRRequest.CsrParams.Organization},
		OrganizationalUnit: []string{genCSRRequest.CsrParams.OrganizationalUnit},
		CommonName:         genCSRRequest.CsrParams.CommonName,
	}

	pemCSR, err := s.manager.GenCSR(subject)
	if err != nil {
		rerr := fmt.Errorf("failed to generate CSR: %v", err)
		log.Error(rerr)
		return rerr
	}

	if err = stream.Send(&pb.RotateCertificateResponse{
		RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{
			GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{
				Type: pb.CertificateType_CT_X509,
				Csr:  pemCSR,
			}},
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

	if loadCertificateRequest.Certificate.Type != pb.CertificateType_CT_X509 {
		rerr := fmt.Errorf("unexpected Certificate type: %d", loadCertificateRequest.Certificate.Type)
		log.Error(rerr)
		return rerr
	}

	certID := loadCertificateRequest.CertificateId
	pemCert := loadCertificateRequest.Certificate.Certificate
	pemCACerts := [][]byte{}
	for _, cert := range loadCertificateRequest.CaCertificate {
		if cert.Type != pb.CertificateType_CT_X509 {
			rerr := fmt.Errorf("unexpected Certificate type: %d", cert.Type)
			log.Error(rerr)
			return rerr
		}
		pemCACerts = append(pemCACerts, cert.Certificate)
	}

	rotateAccept, rotateBack, err := s.manager.Rotate(certID, pemCert, pemCACerts)
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

// EncodeCert encodes a x509.Certificate into a PEM block.
func x509toPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

var certPEMEncoder = x509toPEM

// GetCertificates returns installed certificates.
func (s *Server) GetCertificates(ctx context.Context, request *pb.GetCertificatesRequest) (*pb.GetCertificatesResponse, error) {
	certInfo, err := s.manager.GetCertInfo()
	if err != nil {
		rerr := fmt.Errorf("failed GetCertificates: %v", err)
		log.Error(rerr)
		return nil, rerr
	}
	log.Info("Success GetCertificates.")
	r := []*pb.CertificateInfo{}
	for _, ci := range certInfo {
		r = append(r, &pb.CertificateInfo{
			CertificateId: ci.certID,
			Certificate: &pb.Certificate{
				Type:        pb.CertificateType_CT_X509,
				Certificate: certPEMEncoder(ci.cert),
			},
			ModificationTime: ci.updated.UnixNano(),
		})
	}

	return &pb.GetCertificatesResponse{CertificateInfo: r}, nil
}

// RevokeCertificates revokes certificates.
func (s *Server) RevokeCertificates(ctx context.Context, request *pb.RevokeCertificatesRequest) (*pb.RevokeCertificatesResponse, error) {
	revoked, notRevoked, err := s.manager.Revoke(request.CertificateId)
	if err != nil {
		rerr := fmt.Errorf("failed RevokeCertificates: %v", err)
		log.Error(rerr)
		return nil, rerr
	}

	certRevErr := []*pb.CertificateRevocationError{}
	for certID, errMsg := range notRevoked {
		certRevErr = append(certRevErr, &pb.CertificateRevocationError{
			CertificateId: certID,
			ErrorMessage:  errMsg,
		})
	}

	log.Info("Success RevokeCertificates.")
	return &pb.RevokeCertificatesResponse{RevokedCertificateId: revoked, CertificateRevocationError: certRevErr}, nil
}

// CanGenerateCSR returns if it can generate CSRs with the given properties.
func (s *Server) CanGenerateCSR(ctx context.Context, request *pb.CanGenerateCSRRequest) (*pb.CanGenerateCSRResponse, error) {
	log.Info("Success CanGenerateCSR.")
	ret := &pb.CanGenerateCSRResponse{
		CanGenerate: request.KeyType == pb.KeyType_KT_RSA && request.CertificateType == pb.CertificateType_CT_X509 && request.KeySize >= 128 && request.KeySize <= 4096,
	}
	return ret, nil
}
