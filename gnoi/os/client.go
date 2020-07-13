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
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/gnxi/gnoi/os/pb"
	"github.com/google/gnxi/utils"
	"google.golang.org/grpc"
)

// Client handles requesting OS RPCs.
type Client struct {
	client pb.OSClient
	read   func(string) (file io.ReaderAt, size uint64, err error)
}

const chunkSize = 5000000

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	reader := func(path string) (file io.ReaderAt, size uint64, err error) {
		f, err := os.Open(path)
		if err != nil {
			return
		}
		fileInfo, err := f.Stat()
		if err != nil {
			return
		}
		size = uint64(fileInfo.Size())
		file = f
		return
	}
	return &Client{client: pb.NewOSClient(c), read: reader}
}

// Install invokes the Install RPC for the OS service.
func (c *Client) Install(ctx context.Context, imgPath, version string, printStatus bool, validateTimeout time.Duration) error {
	// Open and Stat OS image.
	if c.read == nil {
		return fmt.Errorf("No reader passed to client")
	}
	file, fileSize, err := c.read(imgPath)

	// Create Install client for streaming.
	install, err := c.client.Install(ctx)
	if err != nil {
		return err
	}
	defer install.CloseSend()

	// Send initial TransferRequest and await response.
	if err = install.Send(&pb.InstallRequest{
		Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: version}},
	}); err != nil {
		return err
	}
	transferResp, err := install.Recv()
	if err != nil {
		return err
	}
	_, validated, err := c.validateInstallRequest(transferResp)
	if err != nil {
		return err
	}
	if validated {
		return fmt.Errorf("OS already installed on target")
	}
	recvErrs := make(chan error)
	recvValidated := make(chan bool, 1)

	// Goroutine to receive responses while sending requests allowing for
	// bidirectional streaming.
	go func() {
		validated := false
		for !validated {
			resp, err := install.Recv()
			if err != nil {
				recvErrs <- err
				continue
			}
			progress, validated, err := c.validateInstallRequest(resp)
			if err != nil {
				recvErrs <- err
				continue
			}
			if validated {
				recvValidated <- validated
				return
			}
			if printStatus {
				fmt.Printf("%d%% transferred\n", progress/fileSize)
			}
		}
	}()

	// Read from file in chunks, sending a chunk of the image each time.
	for n := int64(0); n < int64(fileSize)+int64(chunkSize); n += int64(chunkSize) {
		b := make([]byte, chunkSize)
		if len(recvErrs) > 0 {
			return c.accumulateErrors(recvErrs)
		}
		var readSize int
		if readSize, err = file.ReadAt(b, n); err != nil && err != io.EOF {
			return err
		}
		if readSize == 0 {
			break
		}
		err = install.Send(&pb.InstallRequest{
			Request: &pb.InstallRequest_TransferContent{TransferContent: b},
		})
		if err != nil {
			return err
		}
	}

	// Send TransferEnd to notify targe that last chunk has been transfered.
	if err = install.Send(&pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}}); err != nil {
		return err
	}

	// Await for response from asynchronous receiver or timeout.
	select {
	case <-time.After(validateTimeout):
		return fmt.Errorf("Validation timed out")
	case err = <-recvErrs:
		return fmt.Errorf("%w; %s", c.accumulateErrors(recvErrs), err.Error())
	case <-recvValidated:
		return nil
	}
}

// validateInstallRequest will validate an InstallRequest.
func (c *Client) validateInstallRequest(response *pb.InstallResponse) (progress uint64, validated bool, err error) {
	switch resp := response.Response.(type) {
	case *pb.InstallResponse_Validated:
		validated = true
		return
	case *pb.InstallResponse_TransferProgress:
		progress = resp.TransferProgress.GetBytesReceived()
		return
	case *pb.InstallResponse_InstallError:
		installErr := resp.InstallError
		if installErr.GetType() == pb.InstallError_UNSPECIFIED {
			err = fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
			return
		}
		err = fmt.Errorf("InstallError occured: %s", installErr.GetType().String())
		return
	case *pb.InstallResponse_TransferReady:
		return
	default:
		err = fmt.Errorf("Unexpected response: %T(%v)", resp, resp)
		return
	}
}

// accumulateErrors will deplete errors from a channel and consolidate them.
func (c *Client) accumulateErrors(recvErrs chan error) error {
	err := <-recvErrs
	if err == nil {
		return nil
	}
	for len(recvErrs) > 0 {
		err = fmt.Errorf("%s; %w", (<-recvErrs).Error(), err)
	}
	return err
}

// Activate invokes the Activate RPC for the OS service.
func (c *Client) Activate(ctx context.Context, version string) error {
	request := &pb.ActivateRequest{Version: version}
	utils.LogProto(request)
	response, err := c.client.Activate(ctx, request)
	if err != nil {
		return err
	}
	utils.LogProto(response)
	switch response.Response.(type) {
	case *pb.ActivateResponse_ActivateOk:
		return nil
	case *pb.ActivateResponse_ActivateError:
		res := response.GetActivateError()
		switch res.GetType() {
		case pb.ActivateError_UNSPECIFIED:
			return fmt.Errorf("Unspecified ActivateError: %s", res.GetDetail())
		case pb.ActivateError_NON_EXISTENT_VERSION:
			return fmt.Errorf("Non existent version: %s", version)
		default:
			return fmt.Errorf("Unknown ActivateError: %s", res.GetType())
		}
	}
	return nil
}

// Verify invokes the Verify RPC for the OS service.
func (c *Client) Verify(ctx context.Context) (version, activationFailMsg string, err error) {
	var out *pb.VerifyResponse
	if out, err = c.client.Verify(ctx, &pb.VerifyRequest{}); err != nil {
		return
	}
	version = out.GetVersion()
	activationFailMsg = out.GetActivationFailMessage()
	return
}
