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
	"bytes"
	"context"
	"errors"

	"github.com/google/gnxi/gnoi/os/pb"
	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/mockos"
	"google.golang.org/grpc"
)

const (
	chunkSize = 1000000
)

// Server is an OS Management service.
type Server struct {
	pb.OSServer
	manager     *Manager
	InstallLock chan bool
}

// NewServer returns an OS Management service.
func NewServer(settings *Settings) *Server {
	server := &Server{
		manager:     NewManager(settings.FactoryVersion),
		InstallLock: make(chan bool, 1),
	}
	for _, version := range settings.InstalledVersions {
		server.manager.Install(version, "")
	}
	server.manager.SetRunning(settings.FactoryVersion)
	server.InstallLock <- true
	return server
}


// Register registers the server into the the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterOSServer(g, s)
}

// Activate sets the requested OS version as the version which is used at the next reboot, and reboots the Target.
func (s *Server) Activate(ctx context.Context, request *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	if err := s.manager.SetRunning(request.Version); err != nil {
		return &pb.ActivateResponse{
			Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{
					Type: pb.ActivateError_NON_EXISTENT_VERSION,
				},
			}}, nil
	}
	return &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateOk{}}, nil
}

// Verify returns the OS version currently running.
func (s *Server) Verify(ctx context.Context, _ *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	return &pb.VerifyResponse{
		Version:               s.manager.runningVersion,
		ActivationFailMessage: s.manager.activationFailMessage,
	}, nil
}

// Install receives an OS package, validates the package and then installs the package.
func (s *Server) Install(stream pb.OS_InstallServer) error {
	var request *pb.InstallRequest
	var response *pb.InstallResponse
	var err error
	if request, err = stream.Recv(); err != nil {
		return err
	}
	utils.LogProto(request)
	transferRequest := request.GetTransferRequest()
	if transferRequest == nil {
		return errors.New("Failed to receive TransferRequest")
	}
	if s.manager.IsRunning(transferRequest.Version) {
		response = &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{
			InstallError: &pb.InstallError{Type: pb.InstallError_INSTALL_RUN_PACKAGE},
		}}
		utils.LogProto(response)
		if err = stream.Send(response); err != nil {
			return err
		}
		return errors.New("Attempting to force transfer an OS of the same version as the currently running OS")
	}
	if version := transferRequest.Version; s.manager.IsInstalled(version) {
		response = &pb.InstallResponse{Response: &pb.InstallResponse_Validated{
			Validated: &pb.Validated{
				Version: version,
			},
		}}
		utils.LogProto(response)
		if err = stream.Send(response); err != nil {
			return err
		}
		return nil
	}
	select {
	case <-s.InstallLock:
		defer func() {
			s.InstallLock <- true
		}()
		break
	default:
		response = &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INSTALL_IN_PROGRESS}}}
		utils.LogProto(response)
		if err = stream.Send(response); err != nil {
			return err
		}
		return errors.New("Another install is already in progress")
	}

	response = &pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{TransferReady: &pb.TransferReady{}}}
	utils.LogProto(response)
	if err = stream.Send(response); err != nil {
		return err
	}

	bb, err := ReceiveOS(stream)
	if err != nil {
		return err
	}
	mockOS, err, errResponse := mockos.ValidateOS(bb)
	if err != nil {
		response = &pb.InstallResponse{Response: errResponse}
		utils.LogProto(response)
		if err = stream.Send(response); err != nil {
			return err
		}
		return err
	}
	s.manager.Install(mockOS.Version, mockOS.ActivationFailMessage)
	return nil
}

// ReceiveOS receives and parses requests from stream, storing OS package into a buffer, and updating the progress.
func ReceiveOS(stream pb.OS_InstallServer) (*bytes.Buffer, error) {
	bb := new(bytes.Buffer)
	prev := 0
	for {
		in, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		utils.LogProto(in)
		switch in.Request.(type) {
		case *pb.InstallRequest_TransferContent:
			bb.Write(in.GetTransferContent())
		case *pb.InstallRequest_TransferEnd:
			return bb, nil
		default:
			return nil, errors.New("Unknown request type")
		}
		if curr := bb.Len() / chunkSize; curr > prev {
			prev = curr
			response := &pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{
				TransferProgress: &pb.TransferProgress{BytesReceived: uint64(bb.Len())},
			}}
			utils.LogProto(response)
			if err = stream.Send(response); err != nil {
				return nil, err
			}
		}
	}
}
