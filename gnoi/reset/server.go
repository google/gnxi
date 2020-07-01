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

// Server for factory_reset service.
type Server struct {
	pb.FactoryResetServer
	*Settings
}

// NewServer generates a new factory reset server.
func NewServer(settings *Settings) *Server {
	return &Server{Settings: settings}
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

	// TODO: Reset the target device

	return &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{}}, nil
}
