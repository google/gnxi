package reset

import (
	"context"

	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// Settings for configurable options in Server
type Settings struct {
	errorIfZero   bool
	osUnsupported bool
}

// Server for factory_reset service
type Server struct {
	pb.UnimplementedFactoryResetServer
	*Settings
}

// NewServer generates a new factory reset server
func NewServer(settings *Settings) *Server {
	return &Server{Settings: settings}
}

// Register registers the server into the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterFactoryResetServer(g, s)
}

// Start rpc will start the factory reset process
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	resError := &pb.StartResponse_ResetError{}
	if s.errorIfZero {
		resError.ResetError = &pb.ResetError{ZeroFillUnsupported: true}
	}
	if s.osUnsupported {
		if resError.ResetError != nil {
			resError.ResetError.FactoryOsUnsupported = true
		} else {
			resError.ResetError = &pb.ResetError{FactoryOsUnsupported: true}
		}
	}
	if resError.ResetError != nil {
		return &pb.StartResponse{Response: resError}, nil
	}
	return &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{}}, nil
}
