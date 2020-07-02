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

package reset

import (
	"context"

	cpb "github.com/google/gnxi/gnoi/cert/pb"
	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// CertServerInterface used to pass in cert server for factory reset.
type CertServerInterface interface {
	GetCertificates(context.Context, *cpb.GetCertificatesRequest) (*cpb.GetCertificatesResponse, error)
	RevokeCertificates(context.Context, *cpb.RevokeCertificatesRequest) (*cpb.RevokeCertificatesResponse, error)
}

// Settings for configurable options in Server.
type Settings struct {
	ZeroFillUnsupported  bool
	FactoryOSUnsupported bool
}

// Server for factory_reset service.
type Server struct {
	certServer CertServerInterface
	pb.FactoryResetServer
	*Settings
}

// NewServer generates a new factory reset server.
func NewServer(settings *Settings, certServer CertServerInterface) *Server {
	return &Server{Settings: settings, certServer: certServer}
}

// Register registers the server into the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterFactoryResetServer(g, s)
}

// Start rpc will start the factory reset process.
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	resetError := &pb.ResetError{}
	resetError.ZeroFillUnsupported = req.ZeroFill && s.ZeroFillUnsupported
	resetError.FactoryOsUnsupported = req.FactoryOs && s.FactoryOSUnsupported

	if resetError.ZeroFillUnsupported || resetError.FactoryOsUnsupported {
		return &pb.StartResponse{Response: &pb.StartResponse_ResetError{ResetError: resetError}}, nil
	}

	s.reset()

	return &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{}}, nil
}

// reset the target device. Clears certs and wipes OS's
func (s *Server) reset() error {
	// TODO: Reset the target device using the OS manager when implemented.
	response, err := s.certServer.GetCertificates(context.Background(), &cpb.GetCertificatesRequest{})
	if err != nil {
		return err
	}

	certs := []string{}
	for _, c := range response.CertificateInfo {
		certs = append(certs, c.CertificateId)
	}

	if _, err = s.certServer.RevokeCertificates(context.Background(), &cpb.RevokeCertificatesRequest{CertificateId: certs}); err != nil {
		return err
	}
	return nil
}
