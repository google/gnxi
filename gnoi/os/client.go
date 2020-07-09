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
	"os"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

// Client handles requesting OS RPCs.
type Client struct {
	client pb.OSClient
}

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewOSClient(c)}
}

func (c *Client) Install(ctx context.Context, imgPath, version string) error {
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}
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
	return nil
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
			err = fmt.Errorf("Unspecified error: %s", installErr.GetDetail())
			return
		}
		err = fmt.Errorf("InstallError occured: %s", installErr.GetType().String())
		return
	}
	return
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
