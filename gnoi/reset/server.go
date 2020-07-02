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

	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// Settings for configurable options in Server.
type Settings struct {
	ZeroFillUnsupported  bool
	FactoryOSUnsupported bool
}

// Notifier for reset callback.
type Notifier func()

// Server for factory_reset service.
type Server struct {
	pb.FactoryResetServer
	*Settings
	notifier Notifier
}

// NewServer generates a new factory reset server.
func NewServer(settings *Settings, notifier Notifier) *Server {
	return &Server{Settings: settings, notifier: notifier}
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

	go s.notifier()

	return &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{}}, nil
}
