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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

const readChunkSize = 4000000

type activateRPC func(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error)
type verifyRPC func(ctx context.Context, in *pb.VerifyRequest, opts ...grpc.CallOption) (*pb.VerifyResponse, error)

type mockClient struct {
	pb.OSClient
	activate      activateRPC
	verify        verifyRPC
	installClient pb.OS_InstallClient
}

type installRequestMap struct {
	req  *pb.InstallRequest
	resp *pb.InstallResponse
}

type mockInstallClient struct {
	pb.OS_InstallClient
	reqMap  []*installRequestMap
	i       int
	recv    chan int
	recvErr chan *pb.InstallResponse_InstallError
}

func (c *mockInstallClient) Send(req *pb.InstallRequest) error {
	if c.i < len(c.reqMap) {
		if reflect.TypeOf(req.Request) == reflect.TypeOf(c.reqMap[c.i].req.Request) {
			c.recv <- c.i
		} else {
			c.recvErr <- &pb.InstallResponse_InstallError{
				InstallError: &pb.InstallError{Type: pb.InstallError_UNSPECIFIED, Detail: "Invalid command"},
			}
		}
		c.i++
	}
	return nil
}

func (c *mockInstallClient) Recv() (*pb.InstallResponse, error) {
	select {
	case i := <-c.recv:
		res := c.reqMap[i].resp
		if res == nil {
			<-time.After(200 * time.Millisecond)
			return &pb.InstallResponse{
				Response: &pb.InstallResponse_InstallError{
					InstallError: &pb.InstallError{Type: pb.InstallError_UNSPECIFIED, Detail: "Invalid command"},
				},
			}, nil
		}
		return res, nil
	case err := <-c.recvErr:
		return &pb.InstallResponse{Response: err}, nil
	}
}

func (c *mockInstallClient) CloseSend() error {
	return nil
}

func (c *mockClient) Activate(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
	return c.activate(ctx, in, opts...)
}

func (c *mockClient) Verify(ctx context.Context, in *pb.VerifyRequest, opts ...grpc.CallOption) (*pb.VerifyResponse, error) {
	return c.verify(ctx, in, opts...)
}

func (c *mockClient) Install(ctx context.Context, opts ...grpc.CallOption) (pb.OS_InstallClient, error) {
	return c.installClient, nil
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

func readBytes(num int) func(string) (io.ReaderAt, uint64, func() error, error) {
	b := make([]byte, num)
	rand.Read(b)
	return func(_ string) (io.ReaderAt, uint64, func() error, error) {
		return bytes.NewReader(b), uint64(num), func() error { return nil }, nil
	}
}

func TestInstall(t *testing.T) {
	installTests := []struct {
		name    string
		reqMap  []*installRequestMap
		reader  func(string) (io.ReaderAt, uint64, func() error, error)
		err     error
		timeout time.Duration
	}{
		{
			"Already validated",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_Validated{Validated: &pb.Validated{}}},
				},
			},
			readBytes(0),
			nil,
			100 * time.Millisecond,
		},
		{
			"File size of one chunk then validated",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{TransferProgress: &pb.TransferProgress{BytesReceived: readChunkSize}}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_Validated{Validated: &pb.Validated{}}},
				},
			},
			readBytes(int(readChunkSize)),
			nil,
			100 * time.Millisecond,
		},
		{
			"File size of two chunks + 1 then validated",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{TransferProgress: &pb.TransferProgress{BytesReceived: readChunkSize}}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{TransferProgress: &pb.TransferProgress{BytesReceived: (readChunkSize * 2)}}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{TransferProgress: &pb.TransferProgress{BytesReceived: (readChunkSize * 2) + 1}}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_Validated{Validated: &pb.Validated{}}},
				},
			},
			readBytes((int(readChunkSize) * 2) + 1),
			nil,
			100 * time.Millisecond,
		},
		{
			"File size of one chunk but INCOMPATIBLE InstallError",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INCOMPATIBLE}}},
				},
			},
			readBytes(int(readChunkSize)),
			errors.New("InstallError occurred: INCOMPATIBLE"),
			100 * time.Millisecond,
		},
		{
			"TransferRequest but INCOMPATIBLE InstallError",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INCOMPATIBLE}}},
				},
			},
			readBytes(0),
			errors.New("InstallError occurred: INCOMPATIBLE"),
			100 * time.Millisecond,
		},
		{
			"TransferRequest but Unspecified InstallError",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_UNSPECIFIED, Detail: "Unspecified"}}},
				},
			},
			readBytes(0),
			errors.New("Unspecified InstallError error: Unspecified"),
			100 * time.Millisecond,
		},
		{
			"File size of two chunks but Unspecified InstallError",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_UNSPECIFIED, Detail: "Unspecified"}}},
				},
			},
			readBytes(int(readChunkSize) * 2),
			errors.New("Unspecified InstallError error: Unspecified"),
			100 * time.Millisecond,
		},
		{
			"File size of one chunk then timout",
			[]*installRequestMap{
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: "version"}}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{}},
					&pb.InstallResponse{Response: &pb.InstallResponse_TransferProgress{TransferProgress: &pb.TransferProgress{BytesReceived: readChunkSize}}},
				},
				{
					&pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}},
					nil,
				},
			},
			readBytes(int(readChunkSize)),
			errors.New("Validation timed out"),
			50 * time.Millisecond,
		},
	}
	for _, test := range installTests {
		t.Run(test.name, func(t *testing.T) {
			client := Client{
				client: &mockClient{installClient: &mockInstallClient{
					reqMap:  test.reqMap,
					recv:    make(chan int, 1),
					recvErr: make(chan *pb.InstallResponse_InstallError, 1),
				}},
			}
			fileReader = test.reader
			if err := client.Install(context.Background(), "", "version", test.timeout, readChunkSize); fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("Wanted error: **%v** but got error: **%v**", test.err, err)
			}
		})
	}
}

func TestActivate(t *testing.T) {
	activateTests := []struct {
		name    string
		client  *mockClient
		wantErr bool
	}{
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
	verifyTests := []struct {
		name,
		runningVersion,
		failMessage string
	}{
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
