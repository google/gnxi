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
	"testing"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

type activateRPC func(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error)
type verifyRPC func(ctx context.Context, in *pb.VerifyRequest, opts ...grpc.CallOption) (*pb.VerifyResponse, error)

type mockClient struct {
	pb.OSClient
	activate activateRPC
	verify   verifyRPC
}

type activateTest struct {
	name    string
	client  *mockClient
	wantErr bool
}

type verifyTest struct {
	name,
	runningVersion,
	failMessage string
}

func (c *mockClient) Activate(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
	return c.activate(ctx, in, opts...)
}

func (c *mockClient) Verify(ctx context.Context, in *pb.VerifyRequest, opts ...grpc.CallOption) (*pb.VerifyResponse, error) {
	return c.verify(ctx, in, opts...)
}

func activateErrorRPC(errType pb.ActivateError_Type, detail string) activateRPC {
	return func(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
		return &pb.ActivateResponse{
			Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{Type: errType, Detail: detail},
			}}, nil
	}
}

func activateSuccessRPC(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
	return &pb.ActivateResponse{
		Response: &pb.ActivateResponse_ActivateOk{}}, nil
}

func TestActivate(t *testing.T) {
	activateTests := []activateTest{
		{
			"Success",
			&mockClient{activate: activateSuccessRPC},
			false,
		},
		{
			"Unspecified",
			&mockClient{activate: activateErrorRPC(pb.ActivateError_UNSPECIFIED, "detail")},
			true,
		},
		{
			"Non Existent Version",
			&mockClient{activate: activateErrorRPC(pb.ActivateError_NON_EXISTENT_VERSION, "")},
			true,
		},
	}
	for _, test := range activateTests {
		t.Run(test.name, func(t *testing.T) {
			client := Client{client: test.client}
			got := client.Activate(context.Background(), "version")
			if test.wantErr {
				if got == nil {
					t.Error("want error, got nil")
				}
			} else if got != nil {
				t.Errorf("want <nil>, got: %v", got)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	verifyTests := []verifyTest{
		{"Is Running", "version", ""},
		{"Previous Activation Fail", "version", "Activation fail"},
	}
	for _, test := range verifyTests {
		t.Run(test.name, func(t *testing.T) {
			client := Client{client: &mockClient{
				verify: func(ctx context.Context, in *pb.VerifyRequest, opts ...grpc.CallOption) (*pb.VerifyResponse, error) {
					return &pb.VerifyResponse{Version: test.runningVersion, ActivationFailMessage: test.failMessage}, nil
				},
			}}
			version, activationErr, err := client.Verify(context.Background())
			if err != nil {
				t.Errorf("Verify RPC error: %w", err)
			}
			if activationErr != test.failMessage || version != test.runningVersion {
				t.Errorf(
					"Expected VerifyResponse(%s, %s) but got VerifyResponse(%s, %s)",
					test.runningVersion,
					test.failMessage,
					version,
					activationErr,
				)
			}
		})
	}
}
