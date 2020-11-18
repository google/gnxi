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

type getCertificatesRPC func(ctx context.Context, in *pb.GetCertificatesRequest, opts ...grpc.CallOption) (*pb.GetCertificatesResponse, error)
type revokeCertificatesRPC func(ctx context.Context, in *pb.RevokeCertificatesRequest, opts ...grpc.CallOption) (*pb.RevokeCertificatesResponse, error)
type canGenerateCSRRPC func(ctx context.Context, in *pb.CanGenerateCSRRequest, opts ...grpc.CallOption) (*pb.CanGenerateCSRResponse, error)

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

type installRequestMap struct {
	req  *pb.InstallCertificateRequest
	resp *pb.InstallCertificateResponse
}
type installClient struct {
	pb.CertificateManagement_InstallClient
	reqMap  []*installRequestMap
	i       int
	recv    chan int
	recvErr chan error
}

type mockClient struct {
	pb.CertificateManagementClient
	rotate             *rotateClient
	install            *installClient
	getCertificates    getCertificatesRPC
	revokeCertificates revokeCertificatesRPC
	canGenerateCSR     canGenerateCSRRPC
}

func (c *installClient) Send(req *pb.InstallCertificateRequest) error {
	if c.i < len(c.reqMap) {
		if reflect.TypeOf(req.InstallRequest) == reflect.TypeOf(c.reqMap[c.i].req.InstallRequest) {
			c.recv <- c.i
		} else {
			c.recvErr <- errors.New("error")
		}
		c.i++
	}
	return nil
}

func (c *installClient) Recv() (*pb.InstallCertificateResponse, error) {
	select {
	case i := <-c.recv:
		return c.reqMap[i].resp, nil
	case err := <-c.recvErr:
		return nil, err
	}
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

func (c *mockClient) Install(ctx context.Context, opts ...grpc.CallOption) (pb.CertificateManagement_InstallClient, error) {
	return c.install, nil
}

func (c *mockClient) GetCertificates(ctx context.Context, in *pb.GetCertificatesRequest, opts ...grpc.CallOption) (*pb.GetCertificatesResponse, error) {
	return c.getCertificates(ctx, in, opts...)
}

func (c *mockClient) RevokeCertificates(ctx context.Context, in *pb.RevokeCertificatesRequest, opts ...grpc.CallOption) (*pb.RevokeCertificatesResponse, error) {
	return c.revokeCertificates(ctx, in, opts...)
}

func (c *mockClient) CanGenerateCSR(ctx context.Context, in *pb.CanGenerateCSRRequest, opts ...grpc.CallOption) (*pb.CanGenerateCSRResponse, error) {
	return c.canGenerateCSR(ctx, in, opts...)
}

func TestClientCanGenerateCSR(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			"Will terminate successfully",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{client: &mockClient{canGenerateCSR: func(ctx context.Context, in *pb.CanGenerateCSRRequest, opts ...grpc.CallOption) (*pb.CanGenerateCSRResponse, error) {
				return &pb.CanGenerateCSRResponse{CanGenerate: true}, nil
			}}}
			if _, err := client.CanGenerateCSR(context.Background()); fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted err %v, got err %v", test.err, err)
			}
		})
	}
}

func TestClientRevokeCertificates(t *testing.T) {
	tests := []struct {
		name  string
		wants int
		err   error
	}{
		{
			"Will terminate successfully",
			1,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{client: &mockClient{revokeCertificates: func(ctx context.Context, in *pb.RevokeCertificatesRequest, opts ...grpc.CallOption) (*pb.RevokeCertificatesResponse, error) {
				return &pb.RevokeCertificatesResponse{CertificateRevocationError: []*pb.CertificateRevocationError{{}}}, nil
			}}}
			if _, m, err := client.RevokeCertificates(context.Background(), []string{}); len(m) != test.wants || fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted map of len(%d), got len(%d), wanted err %v, got err %v", test.wants, len(m), test.err, err)
			}
		})
	}
}

func TestClientGetCertificates(t *testing.T) {
	tests := []struct {
		name  string
		wants int
		err   error
	}{{
		"Will terminate successfully",
		1,
		nil,
	},
	}
	PEMtox509 = func(bytes []byte) (*x509.Certificate, error) {
		return &x509.Certificate{}, nil
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{client: &mockClient{getCertificates: func(ctx context.Context, in *pb.GetCertificatesRequest, opts ...grpc.CallOption) (*pb.GetCertificatesResponse, error) {
				return &pb.GetCertificatesResponse{CertificateInfo: []*pb.CertificateInfo{{Certificate: &pb.Certificate{}}}}, nil
			}}}
			if m, err := client.GetCertificates(context.Background()); len(m) != test.wants || fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted map of len(%d), got len(%d), wanted err %v, got err %v", test.wants, len(m), test.err, err)
			}
		})
	}
}

func TestClientInstall(t *testing.T) {
	tests := []struct {
		name     string
		reqMap   []*installRequestMap
		err      error
		sign     func(*x509.CertificateRequest) (*x509.Certificate, error)
		caBundle []*x509.Certificate
	}{
		{
			"Failed to receive InstallCertificateResponse",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: nil},
					&pb.InstallCertificateResponse{},
				},
			},
			errors.New("failed to receive InstallCertificateResponse: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
		},
		{
			"expected GenerateCSRRequest, got something else",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{}},
					&pb.InstallCertificateResponse{InstallResponse: nil},
				},
			},
			errors.New("expected GenerateCSRRequest, got something else"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{},
		},
		{
			"failed to sign the CSR",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{}},
					&pb.InstallCertificateResponse{InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
			},
			errors.New("failed to sign the CSR: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) {
				return nil, errors.New("error")
			},
			[]*x509.Certificate{},
		},
		{
			"failed to receive InstallCertificateResponse",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{}},
					&pb.InstallCertificateResponse{InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.InstallCertificateRequest{InstallRequest: nil},
					&pb.InstallCertificateResponse{InstallResponse: nil},
				},
			},
			errors.New("failed to receive InstallCertificateResponse: error"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{{}},
		},
		{
			"expected LoadCertificateResponse, got something else",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{}},
					&pb.InstallCertificateResponse{InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_LoadCertificate{}},
					&pb.InstallCertificateResponse{InstallResponse: nil},
				},
			},
			errors.New("expected LoadCertificateResponse, got something else"),
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{{}},
		},
		{
			"success",
			[]*installRequestMap{
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_GenerateCsr{}},
					&pb.InstallCertificateResponse{InstallResponse: &pb.InstallCertificateResponse_GeneratedCsr{GeneratedCsr: &pb.GenerateCSRResponse{Csr: &pb.CSR{}}}},
				},
				{
					&pb.InstallCertificateRequest{InstallRequest: &pb.InstallCertificateRequest_LoadCertificate{}},
					&pb.InstallCertificateResponse{InstallResponse: &pb.InstallCertificateResponse_LoadCertificate{LoadCertificate: &pb.LoadCertificateResponse{}}},
				},
			},
			nil,
			func(_ *x509.CertificateRequest) (*x509.Certificate, error) { return &x509.Certificate{}, nil },
			[]*x509.Certificate{{}},
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
			client := &Client{client: &mockClient{install: &installClient{
				reqMap:  test.reqMap,
				recv:    make(chan int, 1),
				recvErr: make(chan error, 1),
			}}}
			if err := client.Install(
				context.Background(), "", 0, pkix.Name{
					Country:            []string{""},
					Organization:       []string{""},
					OrganizationalUnit: []string{""},
					Province:           []string{""},
				}, "", test.sign, test.caBundle,
			); fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted error: **%v** but got error: **%v**", test.err, err)
			}
		})
	}
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
