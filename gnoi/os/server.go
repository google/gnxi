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
	"io"

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
	manager           *Manager
	installInProgress bool
}

// NewServer returns an OS Management service.
func NewServer(settings *Settings) *Server {
	server := &Server{manager: NewManager(settings.FactoryVersion)}
	for _, os := range settings.InstalledVersions {
		server.manager.Install(os)
	}
	return server
}

// Register registers the server into the the gRPC server provided.
func (s *Server) Register(g *grpc.Server) {
	pb.RegisterOSServer(g, s)
}

// Activate sets the requested OS version as the version which is used at the next reboot, and reboots the Target.
func (s *Server) Activate(ctx context.Context, request *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	if err := s.manager.SetRunning(request.Version); err != nil {
		return &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateError{
			ActivateError: &pb.ActivateError{Type: pb.ActivateError_NON_EXISTENT_VERSION},
		}}, nil
	}
	return &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateOk{}}, nil
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
	if s.manager.InstallInProgress() {
		response = &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INSTALL_IN_PROGRESS}}}
		utils.LogProto(response)
		stream.Send(response)
		return errors.New("Another install is already in progress")
	}
	s.manager.SetInstallInProgress(true)
	defer s.manager.SetInstallInProgress(false)

	if version := transferRequest.Version; s.manager.IsInstalled(version) {
		response = &pb.InstallResponse{Response: &pb.InstallResponse_Validated{
			Validated: &pb.Validated{
				Version: version,
			},
		}}
		if err = stream.Send(response); err != nil {
			return err
		}
	}
	response = &pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{TransferReady: &pb.TransferReady{}}}
	if err = stream.Send(response); err != nil {
		return err
	}
	errorChan := make(chan error)
	updateProgress := make(chan uint64)
	transferredOS := make(chan *bytes.Buffer)
	go ReceiveOS(stream, errorChan, updateProgress, transferredOS)

	var bb *bytes.Buffer
streamingProgress:
	for {
		select {
		case err = <-errorChan:
			return err
		case size := <-updateProgress:
			response = &pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{
				TransferProgress: &pb.TransferProgress{BytesReceived: size},
			}}
			utils.LogProto(response)
			if err = stream.Send(response); err != nil {
				return err
			}
		case bb = <-transferredOS:
			break streamingProgress
		}
	}
	mockOS, err, errResponse := mockos.ValidateOS(bb)
	if err != nil {
		stream.Send(&pb.InstallResponse{Response: errResponse})
		return err
	}
	s.manager.Install(mockOS.Version, mockOS.ActivationFailMessage)
	return nil
}

// ReceiveOS receives and parses requests from stream, storing OS package into a buffer.
func ReceiveOS(stream pb.OS_InstallServer, errorChan chan error, updateProgress chan uint64, transferredOS chan *bytes.Buffer) {
	bb := new(bytes.Buffer)
	prev := 0
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			errorChan <- nil
			return
		}
		if err != nil {
			errorChan <- err
			return
		}
		utils.LogProto(in)
		switch in.Request.(type) {
		case *pb.InstallRequest_TransferContent:
			bb.Write(in.GetTransferContent())
		case *pb.InstallRequest_TransferEnd:
			transferredOS <- bb
			errorChan <- nil
			return
		default:
			errorChan <- errors.New("Unknown request type")
			return
		}
		if curr := bb.Len() / chunkSize; curr > prev {
			prev = curr
			updateProgress <- uint64(bb.Len())
		}
	}
}
