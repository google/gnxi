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

package os

import (
	"context"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

// Server is an OS Management service.
type Server struct {
	pb.OSServer
	manager               *Manager
	activationFailMessage string
}

// NewServer returns an OS Management service.
func NewServer(settings *Settings) *Server {
	server := &Server{
		manager: NewManager(settings.FactoryVersion),
	}
	for _, os := range settings.InstalledVersions {
		server.manager.Install(os)
	}
	server.manager.SetRunning(settings.FactoryVersion)
	return server
}

// Register registers the server into the the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterOSServer(g, s)
}

// Activate sets the requested OS version as the version which is used at the next reboot, and reboots the Target.
func (s *Server) Activate(ctx context.Context, request *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	if msg := s.manager.failMsgs[request.Version]; msg != "" {
		s.activationFailMessage = msg
	} else if err := s.manager.SetRunning(request.Version); err != nil {
		return &pb.ActivateResponse{
			Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{
					Type: pb.ActivateError_NON_EXISTENT_VERSION,
				},
			}}, nil
	} else {
		s.activationFailMessage = ""
	}
	return &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateOk{}}, nil
}

// Verify returns the OS version currently running.
func (s *Server) Verify(ctx context.Context, _ *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	return &pb.VerifyResponse{
		Version:               s.manager.runningVersion,
		ActivationFailMessage: s.activationFailMessage,
	}, nil
}
