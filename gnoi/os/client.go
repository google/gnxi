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

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

// Client handles requesting OS RPCs.
type Client struct {
	client pb.OSClient
}

// ActivateErrorType represents the Type enum in ActivateError.
type ActivateErrorType string

// Enum representing possible ActivateErrorTypes.
const (
	ActivateUnspecified        ActivateErrorType = "UNSPECIFIED"
	ActivateNonExistentVersion                   = "NON_EXISTENT_VERSION"
)

// ActivateError represents an error returned by the Activate RPC.
type ActivateError struct {
	error
	ErrType ActivateErrorType
	Detail  string
}

func (e *ActivateError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrType, e.Detail)
}

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewOSClient(c)}
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
		errType := ActivateErrorType(res.GetType().String())
		switch errType {
		case ActivateUnspecified:
			fallthrough
		case ActivateNonExistentVersion:
			return &ActivateError{
				ErrType: errType,
				Detail:  res.GetDetail(),
			}
		default:
			return fmt.Errorf("Unknown ActivateError type: %s", errType)
		}
	}
	return nil
}
