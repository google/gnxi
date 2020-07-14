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

	log "github.com/golang/glog"
	"github.com/google/gnxi/gnoi/os/pb"
	"github.com/google/gnxi/utils"
	"google.golang.org/grpc"
)

var fileReader = func(path string) (file io.ReaderAt, size uint64, close func() error, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}
	var fileInfo os.FileInfo
	fileInfo, err = f.Stat()
	if err != nil {
		return
	}
	size = uint64(fileInfo.Size())
	file = f
	close = f.Close
	return
}

// Client handles requesting OS RPCs.
type Client struct {
	client pb.OSClient
}

const chunkSize = 5000000

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewOSClient(c)}
}

// Install invokes the Install RPC for the OS service.
func (c *Client) Install(ctx context.Context, imgPath, version string, printStatus bool, validateTimeout time.Duration) error {
	file, fileSize, fileClose, err := fileReader(imgPath)
	if err != nil {
		return err
	}
	defer fileClose()

	cancelCtx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()

	install, err := c.client.Install(cancelCtx)
	if err != nil {
		return err
	}

	// Send initial TransferRequest and await response.
	request := &pb.InstallRequest{
		Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: version}},
	}
	utils.LogProto(request)
	if err = install.Send(request); err != nil {
		return err
	}
	var transferResp *pb.InstallResponse

	if transferResp, err = install.Recv(); err != nil {
		return err
	}
	utils.LogProto(transferResp)
	switch resp := transferResp.Response.(type) {
	case *pb.InstallResponse_Validated:
		log.Info("OS already validated")
		return nil
	case *pb.InstallResponse_InstallError:
		installErr := resp.InstallError
		if installErr.GetType() == pb.InstallError_UNSPECIFIED {
			return fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
		}
		return fmt.Errorf("InstallError occured: %s", installErr.GetType().String())
	case *pb.InstallResponse_TransferReady:
	default:
		return fmt.Errorf("Unexpected response: %T(%v)", resp, resp)
	}

	errs := make(chan error, 2)
	validated := make(chan bool, 1)
	doneSend := make(chan bool, 1)

	// Goroutine to receive responses while sending requests allowing for
	// bidirectional streaming.
	go func() {
		for {
			response, err := install.Recv()
			if err != nil {
				errs <- err
				return
			}
			switch resp := response.Response.(type) {
			case *pb.InstallResponse_TransferProgress:
				if printStatus {
					fmt.Printf("%d%% transferred\n", resp.TransferProgress.GetBytesReceived()/fileSize)
				}
			case *pb.InstallResponse_Validated:
				utils.LogProto(response)
				validated <- true
				return
			case *pb.InstallResponse_InstallError:
				utils.LogProto(response)
				installErr := resp.InstallError
				if installErr.GetType() == pb.InstallError_UNSPECIFIED {
					err = fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
					errs <- err
					return
				}
				err = fmt.Errorf("InstallError occured: %s", installErr.GetType().String())
				errs <- err
				return
			default:
				utils.LogProto(response)
				err = fmt.Errorf("Unexpected response: %T(%v)", resp, resp)
				errs <- err
				return
			}
		}
	}()

	// Goroutine to read from file in chunks, sending a chunk of the
	// image each time.
	go func() {
		var readSize int
		b := make([]byte, chunkSize)
		for n := int64(0); n < int64(fileSize)+int64(chunkSize); n += int64(chunkSize) {
			if readSize, err = file.ReadAt(b, n); err != nil && err != io.EOF {
				errs <- err
				return
			}
			if readSize == 0 {
				break
			}
			if err = install.Send(&pb.InstallRequest{
				Request: &pb.InstallRequest_TransferContent{TransferContent: b[:readSize]},
			}); err != nil {
				errs <- err
				return
			}
		}
		doneSend <- true
	}()

	// Await for response from asynchronous receiver or timeout.
	select {
	case <-doneSend:
		request = &pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}}
		utils.LogProto(request)
		if err = install.Send(request); err != nil {
			return err
		}
	case err := <-errs:
		return err
	}

	select {
	case <-time.After(validateTimeout):
		return fmt.Errorf("Validation timed out")
	case err = <-errs:
		return err
	case <-validated:
	}
	return nil
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
