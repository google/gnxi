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
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

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
	file, err := ioutil.ReadFile(imgPath)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(file)
	install, err := c.client.Install(ctx)
	defer install.CloseSend()
	if err != nil {
		return err
	}

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
				fmt.Printf("%d%% transferred\n", progress)
			}
		}
	}()
	for buffer.Len() > 0 {
		b := make([]byte, chunkSize)
		if len(recvErrs) > 0 {
			return c.accumulateErrors(recvErrs)
		}
		_, err = buffer.Read(b)
		if err != nil && err != io.EOF {
			return err
		}
		err = install.Send(&pb.InstallRequest{
			Request: &pb.InstallRequest_TransferContent{TransferContent: b},
		})
		if err != nil {
			return err
		}
	}

	select {
	case <-time.After(validateTimeout):
		return fmt.Errorf("Validation timed out")
	case err = <-recvErrs:
		return fmt.Errorf("%w; %s", c.accumulateErrors(recvErrs), err.Error())
	case <-recvValidated:
		return nil
	}
}

func (c *Client) validateInstallRequest(response *pb.InstallResponse) (progress uint32, validated bool, err error) {
	switch resp := response.Response.(type) {
	case *pb.InstallResponse_Validated:
		validated = true
		return
	case *pb.InstallResponse_SyncProgress:
		progress = resp.SyncProgress.GetPercentageTransferred()
		return
	case *pb.InstallResponse_InstallError:
		installErr := resp.InstallError
		if installErr.GetType() == pb.InstallError_UNSPECIFIED {
			err = fmt.Errorf("Unspecified InstallError error: %s", installErr.GetDetail())
			return
		}
		err = fmt.Errorf("InstallError occured: %s", installErr.GetType().String())
		return
	}
	return
}

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
	out, err := c.client.Activate(ctx, &pb.ActivateRequest{Version: version})
	if err != nil {
		return err
	}
	switch out.Response.(type) {
	case *pb.ActivateResponse_ActivateOk:
		return nil
	case *pb.ActivateResponse_ActivateError:
		res := out.GetActivateError()
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
