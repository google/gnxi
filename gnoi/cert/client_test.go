/* Copyright 2020 Google Inc.

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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/gnxi/gnoi/cert/pb"
	"google.golang.org/grpc"
)

type rotateRequestMap struct {
	req  *pb.RotateCertificateRequest
	resp *pb.RotateCertificateResponse
}
type rotateClient struct {
	pb.CertificateManagement_RotateClient
	reqMap  []*rotateRequestMap
	i       int
	recv    chan int
	recvErr chan error
}

type mockClient struct {
	pb.CertificateManagementClient
	rotate *rotateClient
}

func (c *rotateClient) Send(req *pb.RotateCertificateRequest) error {
	if c.i < len(c.reqMap) {
		if reflect.TypeOf(req.RotateRequest) == reflect.TypeOf(c.reqMap[c.i].req.RotateRequest) {
			c.recv <- c.i
		} else {
			c.recvErr <- errors.New("error")
		}
		c.i++
	}
	return nil
}

func (c *rotateClient) Recv() (*pb.RotateCertificateResponse, error) {
	select {
	case i := <-c.recv:
		return c.reqMap[i].resp, nil
	case err := <-c.recvErr:
		return nil, err
	}
}

func (c *mockClient) Rotate(ctx context.Context, opts ...grpc.CallOption) (pb.CertificateManagement_RotateClient, error) {
	return c.rotate, nil
}

func TestClientRotate(t *testing.T) {
	tests := []struct {
		name     string
		reqMap   []*rotateRequestMap
		err      error
		sign     func(*x509.CertificateRequest) (*x509.Certificate, error)
		caBundle []*x509.Certificate
		validate func() error
	}{
		{
			"Failed to receive RotateCertificateResponse",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: nil},
					&pb.RotateCertificateResponse{},
				},
			},
			errors.New("failed to receive RotateCertificateResponse: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
			func() error { return nil },
		},
		{
			"Expected GenerateCSRRequest",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: nil},
				},
			},
			errors.New("expected GenerateCSRRequest, got something else"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
			func() error { return nil },
		},
		{
			"Fail to sign the CSR",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
			},
			errors.New("failed to sign the CSR: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) {
				return nil, errors.New("error")
			},
			[]*x509.Certificate{},
			func() error { return nil },
		},
		{
			"Failed to receive RotateCertificateResponse",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.RotateCertificateRequest{RotateRequest: nil},
					nil,
				},
			},
			errors.New("failed to receive RotateCertificateResponse: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
			func() error { return nil },
		},
		{
			"expected LoadCertificateResponse, got something else",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{}},
					&pb.RotateCertificateResponse{RotateResponse: nil},
				},
			},
			errors.New("expected LoadCertificateResponse, got something else"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
			func() error { return nil },
		},
		{
			"Validation error",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_LoadCertificate{LoadCertificate: &pb.LoadCertificateResponse{}}},
				},
			},
			errors.New("failed to validate rotated certificate: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
			func() error { return errors.New("error") },
		},
		{
			"Successful",
			[]*rotateRequestMap{
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_GenerateCsr{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_LoadCertificate{}},
					&pb.RotateCertificateResponse{RotateResponse: &pb.RotateCertificateResponse_LoadCertificate{LoadCertificate: &pb.LoadCertificateResponse{}}},
				},
				{
					&pb.RotateCertificateRequest{RotateRequest: &pb.RotateCertificateRequest_FinalizeRotation{FinalizeRotation: &pb.FinalizeRequest{}}},
					nil,
				},
			},
			nil,
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{{}},
			func() error { return nil },
		},
	}
	x509toPEM = func(cert *x509.Certificate) []byte {
		return []byte{}
	}
	parseCSR = func(genCSR *pb.GenerateCSRResponse) (*x509.CertificateRequest, error) {
		return &x509.CertificateRequest{}, nil
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{client: &mockClient{rotate: &rotateClient{
				reqMap:  test.reqMap,
				recv:    make(chan int, 1),
				recvErr: make(chan error, 1),
			}}}
			if err := client.Rotate(
				context.Background(), "", 0, pkix.Name{
					Country:            []string{""},
					Organization:       []string{""},
					OrganizationalUnit: []string{""},
					Province:           []string{""},
				}, "", test.sign, test.caBundle, test.validate,
			); fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted error: **%v** but got error: **%v**", test.err, err)
			}
		})
	}
}
