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
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"

	"github.com/google/gnxi/gnoi/cert/pb"
	"github.com/google/gnxi/utils"
	"google.golang.org/grpc"
)

// Client is a Certificate Management service client.
type Client struct {
	client pb.CertificateManagementClient
}

// NewClient returns a new Client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewCertificateManagementClient(c)}
}

var parseCSR = func(genCSR *pb.GenerateCSRResponse) (*x509.CertificateRequest, error) {
	derCSR, _ := pem.Decode(genCSR.Csr.Csr)
	if derCSR == nil {
		return nil, fmt.Errorf("failed to decode CSR PEM block")
	}

	csr, err := x509.ParseCertificateRequest(derCSR.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR DER")
	}
	return csr, nil
}

// Rotate rotates a certificate.
func (c *Client) Rotate(ctx context.Context, certID string, minKeySize uint32, params pkix.Name, ipAddress string, sign func(*x509.CertificateRequest) (*x509.Certificate, error), caBundle []*x509.Certificate, validate func() error) error {
	stream, err := c.client.Rotate(ctx)
	if err != nil {
		return fmt.Errorf("failed stream: %v", err)
	}
	request := &pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{
			GenerateCsr: &pb.GenerateCSRRequest{
				CsrParams: &pb.CSRParams{
					Type:               pb.CertificateType_CT_X509,
					MinKeySize:         minKeySize,
					KeyType:            pb.KeyType_KT_RSA,
					CommonName:         params.CommonName,
					Country:            params.Country[0],
					Organization:       params.Organization[0],
					OrganizationalUnit: params.OrganizationalUnit[0],
					State:              params.Province[0],
					IpAddress:          ipAddress,
				},
				CertificateId: certID,
			},
		},
	}
	utils.LogProto(request)
	if err = stream.Send(request); err != nil {
		return fmt.Errorf("failed to send GenerateCSRRequest: %v", err)
	}
	var response *pb.RotateCertificateResponse
	if response, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateResponse: %v", err)
	}
	utils.LogProto(response)
	genCSR := response.GetGeneratedCsr()
	if genCSR == nil || genCSR.Csr == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}

	csr, err := parseCSR(genCSR)
	if err != nil {
		return err
	}

	signedCert, err := sign(csr)
	if err != nil {
		return fmt.Errorf("failed to sign the CSR: %v", err)
	}

	certPEM := x509toPEM(signedCert)

	caCertificates := []*pb.Certificate{}
	for _, caCert := range caBundle {
		caCertificates = append(caCertificates, &pb.Certificate{
			Type:        pb.CertificateType_CT_X509,
			Certificate: x509toPEM(caCert),
		})
	}
	request = &pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateRequest{
				Certificate: &pb.Certificate{
					Type:        pb.CertificateType_CT_X509,
					Certificate: certPEM,
				},
				CaCertificates: caCertificates,
			},
		},
	}
	utils.LogProto(request)
	if err = stream.Send(request); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}
	if response, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive RotateCertificateResponse: %v", err)
	}
	utils.LogProto(response)
	loadCertificateResponse := response.GetLoadCertificate()
	if loadCertificateResponse == nil {
		return fmt.Errorf("expected LoadCertificateResponse, got something else")
	}

	if err := validate(); err != nil {
		return fmt.Errorf("failed to validate rotated certificate: %v", err)
	}

	request = &pb.RotateCertificateRequest{
		RotateRequest: &pb.RotateCertificateRequest_FinalizeRotation{FinalizeRotation: &pb.FinalizeRequest{}},
	}
	utils.LogProto(request)
	if err := stream.Send(request); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}
	return nil
}

// Install installs a certificate.
func (c *Client) Install(ctx context.Context, certID string, minKeySize uint32, params pkix.Name, ipAddress string, sign func(*x509.CertificateRequest) (*x509.Certificate, error), caBundle []*x509.Certificate) error {
	stream, err := c.client.Install(ctx)
	if err != nil {
		return fmt.Errorf("failed stream: %v", err)
	}
	request := &pb.InstallCertificateRequest{
		InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{
			GenerateCsr: &pb.GenerateCSRRequest{CsrParams: &pb.CSRParams{
				Type:               pb.CertificateType_CT_X509,
				MinKeySize:         minKeySize,
				KeyType:            pb.KeyType_KT_RSA,
				CommonName:         params.CommonName,
				Country:            params.Country[0],
				Organization:       params.Organization[0],
				OrganizationalUnit: params.OrganizationalUnit[0],
				State:              params.Province[0],
				IpAddress:          ipAddress,
			},
				CertificateId: certID,
			},
		},
	}
	utils.LogProto(request)
	if err = stream.Send(request); err != nil {
		return fmt.Errorf("failed to send GenerateCSRRequest: %v", err)
	}

	var response *pb.InstallCertificateResponse
	if response, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateResponse: %v", err)
	}
	utils.LogProto(response)

	genCSR := response.GetGeneratedCsr()
	if genCSR == nil || genCSR.Csr == nil {
		return fmt.Errorf("expected GenerateCSRRequest, got something else")
	}

	csr, err := parseCSR(genCSR)
	if err != nil {
		return err
	}

	signedCert, err := sign(csr)
	if err != nil {
		return fmt.Errorf("failed to sign the CSR: %v", err)
	}

	certPEM := x509toPEM(signedCert)

	caCertificates := []*pb.Certificate{}
	for _, caCert := range caBundle {
		caCertificates = append(caCertificates, &pb.Certificate{
			Type:        pb.CertificateType_CT_X509,
			Certificate: x509toPEM(caCert),
		})
	}

	request = &pb.InstallCertificateRequest{
		InstallRequest: &pb.InstallCertificateRequest_LoadCertificate{
			LoadCertificate: &pb.LoadCertificateRequest{
				Certificate: &pb.Certificate{
					Type:        pb.CertificateType_CT_X509,
					Certificate: certPEM,
				},
				CaCertificates: caCertificates,
			},
		},
	}
	utils.LogProto(request)
	if err = stream.Send(request); err != nil {
		return fmt.Errorf("failed to send LoadCertificateRequest: %v", err)
	}

	if response, err = stream.Recv(); err != nil {
		return fmt.Errorf("failed to receive InstallCertificateResponse: %v", err)
	}
	utils.LogProto(response)
	loadCertificateResponse := response.GetLoadCertificate()
	if loadCertificateResponse == nil {
		return fmt.Errorf("expected LoadCertificateResponse, got something else")
	}
	return nil
}

// GetCertificates gets a map of certificates in the target, certID to certificate
func (c *Client) GetCertificates(ctx context.Context) (map[string]*x509.Certificate, error) {
	request := &pb.GetCertificatesRequest{}
	utils.LogProto(request)
	response, err := c.client.GetCertificates(ctx, request)
	if err != nil {
		return nil, err
	}
	utils.LogProto(response)
	ret := map[string]*x509.Certificate{}
	for _, certInfo := range response.CertificateInfo {
		if certInfo.Certificate == nil {
			continue
		}
		x509Cert, err := PEMtox509(certInfo.Certificate.Certificate)
		if err != nil {
			return nil, fmt.Errorf("failed to decode certificate: %v", err)
		}
		ret[certInfo.CertificateId] = x509Cert
	}
	return ret, nil
}

// RevokeCertificates revokes certificates in the target, returns a map of certID to error for the ones that failed to be revoked.
func (c *Client) RevokeCertificates(ctx context.Context, certIDs []string) (map[string]string, error) {
	request := &pb.RevokeCertificatesRequest{CertificateId: certIDs}
	utils.LogProto(request)
	response, err := c.client.RevokeCertificates(ctx, request)
	if err != nil {
		return nil, err
	}
	utils.LogProto(response)
	ret := map[string]string{}
	for _, revError := range response.CertificateRevocationError {
		ret[revError.CertificateId] = revError.ErrorMessage
	}
	return ret, nil
}

// CanGenerateCSR checks if the target can generate a CSR.
func (c *Client) CanGenerateCSR(ctx context.Context) (bool, error) {
	request := &pb.CanGenerateCSRRequest{
		KeyType:         pb.KeyType_KT_RSA,
		CertificateType: pb.CertificateType_CT_X509,
		KeySize:         2048,
	}
	utils.LogProto(request)
	response, err := c.client.CanGenerateCSR(ctx, request)
	if err != nil {
		return false, err
	}
	utils.LogProto(response)
	return response.CanGenerate, nil
}
