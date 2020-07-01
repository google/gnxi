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

package reset

import (
	"context"
	"errors"
	"log"

	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// Client handles requesting a Factory Reset.
type Client struct {
	client pb.FactoryResetClient
}

// NewClient initializes a FactoryReset Client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewFactoryResetClient(c)}
}

// ResetTarget invokes gRPC start service on the server.
func (c *Client) ResetTarget(ctx context.Context, zeroFill, rollbackOS bool) error {
	out, err := c.client.Start(ctx, &pb.StartRequest{
		FactoryOs: rollbackOS,
		ZeroFill:  zeroFill,
	})
	if err != nil {
		log.Println("Error calling Start service:", err)
		return err
	}
	return CheckResponse(out)
}

// CheckResponse checks for errors.
func CheckResponse(res *pb.StartResponse) error {
	switch res.Response.(type) {
	case *pb.StartResponse_ResetSuccess:
		return nil
	case *pb.StartResponse_ResetError:
		resErr := res.GetResetError()
		if resErr.FactoryOsUnsupported {
			log.Println("Factory OS Rollback Unsupported")
			return errors.New("Factory OS Rollback Unsupported")
		}
		if resErr.ZeroFillUnsupported {
			log.Println("Zero Filling Persistent Storage Unsupported")
			return errors.New("Zero Fill Unsupported")
		}
		if resErr.Other {
			log.Println(resErr.Detail)
			return errors.New(resErr.Detail)
		}
	}
	return nil
}
