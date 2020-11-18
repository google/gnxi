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
	"strings"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// Client handles requesting a Factory Reset.
type Client struct {
	client pb.FactoryResetClient
}

// ResetError allows the return of multiple error messages concatenated.
type ResetError struct {
	Msgs []string
}

// Error concatenates a multi-line error message.
func (re *ResetError) Error() string {
	return strings.Join(re.Msgs, "\n")
}

// NewClient initializes a FactoryReset Client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewFactoryResetClient(c)}
}

// ResetTarget invokes gRPC start service on the server.
func (c *Client) ResetTarget(ctx context.Context, zeroFill, rollbackOS bool) *ResetError {
	request := &pb.StartRequest{
		FactoryOs: rollbackOS,
		ZeroFill:  zeroFill,
	}
	log.V(1).Info("StartRequest:\n", proto.MarshalTextString(request))
	response, err := c.client.Start(ctx, request)
	if err != nil {
		return &ResetError{Msgs: []string{err.Error()}}
	}
	log.V(1).Info("StartResponse:\n", proto.MarshalTextString(response))
	return CheckResponse(response)
}

// CheckResponse checks for errors.
func CheckResponse(res *pb.StartResponse) *ResetError {
	switch res.Response.(type) {
	case *pb.StartResponse_ResetSuccess:
		return nil
	case *pb.StartResponse_ResetError:
		resErr := res.GetResetError()
		err := &ResetError{
			Msgs: make([]string, 0),
		}
		if resErr.FactoryOsUnsupported {
			err.Msgs = append(err.Msgs, "Factory OS Rollback Unsupported")
		}
		if resErr.ZeroFillUnsupported {
			err.Msgs = append(err.Msgs, "Zero Filling Persistent Storage Unsupported")
		}
		if resErr.Other {
			err.Msgs = append(err.Msgs, "Unspecified Error: "+resErr.Detail)
		}
		return err
	}
	return nil
}
