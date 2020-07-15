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
	"crypto/rand"
	"errors"
	"testing"

	"github.com/google/gnxi/gnoi/os/pb"
	"github.com/kylelemons/godebug/pretty"
)

type mockStream struct {
	pb.OS_InstallServer
	responses chan *pb.InstallResponse
	errorReq  *pb.InstallRequest
}

func (m mockStream) Send(res *pb.InstallResponse) error {
	m.responses <- res
	return nil
}

func (m mockStream) Recv() (*pb.InstallRequest, error) {
	if request := m.errorReq; request != nil {
		return request, nil
	}
	select {
	case <-m.responses:
		return &pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}}, nil
	default:
		buf := make([]byte, 1000000)
		rand.Read(buf)
		return &pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{TransferContent: buf}}, nil
	}
}

func TestTargetActivate(t *testing.T) {
	settings := &Settings{
		FactoryVersion:    "1",
		InstalledVersions: []string{"1.0.0a", "2.0.1c"},
	}
	server := NewServer(settings)
	tests := []struct {
		request *pb.ActivateRequest
		want    *pb.ActivateResponse
	}{
		{
			request: &pb.ActivateRequest{
				Version: "1.0.0a",
			},
			want: &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateOk{}},
		},
		{
			request: &pb.ActivateRequest{
				Version: "99.0a",
			},
			want: &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{Type: pb.ActivateError_NON_EXISTENT_VERSION},
			}},
		},
	}
	for _, test := range tests {
		got, _ := server.Activate(context.Background(), test.request)
		if diff := pretty.Compare(test.want.Response, got.Response); diff != "" {
			t.Errorf("Activate(context.Background(), %s): (-want +got):\n%s", test.request, diff)
		}
	}
}

func TestTargetVerify(t *testing.T) {
	tests := []struct {
		settings *Settings
		want     *pb.VerifyResponse
	}{
		{
			settings: &Settings{
				FactoryVersion: "1",
			},
			want: &pb.VerifyResponse{
				Version: "1",
			},
		},
	}
	for _, test := range tests {
		server := NewServer(test.settings)
		got, _ := server.Verify(context.Background(), &pb.VerifyRequest{})
		if diff := pretty.Compare(test.want, got); diff != "" {
			t.Errorf("Verify(context.Background(), &pb.VerifyRequest{}): (-want +got):\n%s", diff)
		}
	}
}

func TestTargetVerifyFail(t *testing.T) {
	tests := []struct {
		settings *Settings
		want     *pb.VerifyResponse
	}{
		{
			settings: &Settings{
				FactoryVersion: "1",
			},
			want: &pb.VerifyResponse{
				Version:               "1",
				ActivationFailMessage: "Failed to activate OS...",
			},
		},
	}
	for _, test := range tests {
		server := NewServer(test.settings)
		server.manager.activationFailMessage = "Failed to activate OS..."
		got, _ := server.Verify(context.Background(), &pb.VerifyRequest{})
		if diff := pretty.Compare(test.want, got); diff != "" {
			t.Errorf("Verify(context.Background(), &pb.VerifyRequest{}): (-want +got):\n%s", diff)
		}
	}
}

func TestTargetActivateAndVerify(t *testing.T) {
	test := struct {
		settings *Settings
		want     *pb.VerifyResponse
	}{
		settings: &Settings{
			FactoryVersion: "1",
		},
		want: &pb.VerifyResponse{
			Version:               "1",
			ActivationFailMessage: "Failed to activate OS...",
		},
	}
	server := NewServer(test.settings)
	server.manager.Install("1.0.1a", "Failed to activate OS...")
	server.Activate(context.Background(), &pb.ActivateRequest{Version: "1.0.1a"})
	got, _ := server.Verify(context.Background(), &pb.VerifyRequest{})
	if diff := pretty.Compare(test.want, got); diff != "" {
		t.Errorf("Verify(context.Background(), &pb.VerifyRequest{}): (-want +got):\n%s", diff)
	}
}

func TestTargetReceiveOS(t *testing.T) {
	tests := []struct {
		stream *mockStream
		err    error
	}{
		{
			stream: &mockStream{
				responses: make(chan *pb.InstallResponse, 1),
			},
			err: nil,
		},
		{
			stream: &mockStream{
				responses: make(chan *pb.InstallResponse, 1),
				errorReq:  &pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{}}, // Unexpected request after transfer begins.
			},
			err: errors.New("Unknown request type"),
		},
	}
	for _, test := range tests {
		_, err := ReceiveOS(test.stream)
		if diff := pretty.Compare(test.err, err); diff != "" {
			t.Errorf("ReceiveOS(stream): (-want +got):\n%s", diff)
		}
	}
}
