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
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

var (
	printProgess = flag.Bool("print_progress", false, "Prints progress periodically of file transfer.")
)

func fileReader(path string) (file io.ReaderAt, size uint64, close func() error, err error) {
	// Clean the path to remove any ../ sequences
	cleanPath := filepath.Clean(path)

	// Reject any path that tries to escape (contains ..)
	// Also reject absolute paths to prevent reading arbitrary system files
	if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return nil, 0, nil, fmt.Errorf("invalid path: %s", path)
	}

	var f *os.File
	f, err = os.Open(cleanPath)
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

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewOSClient(c)}
}

// Install invokes the Install RPC for the OS service.
func (c *Client) Install(ctx context.Context, imgPath, version string, validateTimeout time.Duration, chunkSize uint64) error {
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
	log.V(1).Info("InstallRequest:\n", proto.MarshalTextString(request))
	if err = install.Send(request); err != nil {
		return err
	}

	var transferResp *pb.InstallResponse
	if transferResp, err = install.Recv(); err != nil {
		return err
	}
	log.V(1).Info("InstallResponse:\n", proto.MarshalTextString(transferResp))
	switch resp := transferResp.Response.(type) {
	case *pb.InstallResponse_Validated:
		log.Infof("OS version %s is already installed", version)
		return nil
	case *pb.InstallResponse_InstallError:
		installErr := resp.InstallError
		if installErr.GetType() == pb.InstallError_UNSPECIFIED {
			return fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
		}
		return fmt.Errorf("InstallError occurred: %s", installErr.GetType().String())
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
				if *printProgess {
					fmt.Printf("%d%% transferred\n", resp.TransferProgress.GetBytesReceived()/fileSize)
				}
			case *pb.InstallResponse_Validated:
				log.V(1).Info("InstallResponse_Validated:\n", proto.MarshalTextString(response))
				validated <- true
				return
			case *pb.InstallResponse_InstallError:
				log.V(1).Info("InstallResponse_InstallError:\n", proto.MarshalTextString(response))
				installErr := resp.InstallError
				if installErr.GetType() == pb.InstallError_UNSPECIFIED {
					err = fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
					errs <- err
					return
				}
				err = fmt.Errorf("InstallError occurred: %s", installErr.GetType().String())
				errs <- err
				return
			default:
				log.V(1).Info("Unexpected:\n", proto.MarshalTextString(response))
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
			if readSize, err = file.ReadAt(b, n); err != nil {
				if err == io.EOF {
					break
				}
				errs <- err
				return
			}
			transferContent := &pb.InstallRequest{
				Request: &pb.InstallRequest_TransferContent{TransferContent: b[:readSize]},
			}
			log.V(1).Info("InstallRequest:\n", proto.MarshalTextString(transferContent))
			if err = install.Send(transferContent); err != nil {
				errs <- err
				return
			}
		}
		transferEnd := &pb.InstallRequest{
			Request: &pb.InstallRequest_TransferEnd{TransferEnd: &pb.TransferEnd{}},
		}
		log.V(1).Info("InstallRequest:\n", proto.MarshalTextString(transferEnd))
		if err = install.Send(transferEnd); err != nil {
			errs <- err
			return
		}
		doneSend <- true
	}()

	select {
	case err = <-errs:
		return err
	case <-doneSend:
	}
	select {
	case err = <-errs:
		return err
	case <-validated:
		return nil
	case <-time.After(validateTimeout):
		return fmt.Errorf("Install timed out")
	}
}
