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

	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/gnoi/os/pb"
	"github.com/google/gnxi/utils/mockos"
	mockosPb "github.com/google/gnxi/utils/mockos/pb"
	"github.com/kylelemons/godebug/pretty"
)

type installResult struct {
	res *pb.InstallResponse
	err error
}

type mockTransferStream struct {
	pb.OS_InstallServer
	response chan *pb.InstallResponse
	errorReq *pb.InstallRequest
	result   chan *pb.InstallResponse
	os       *mockos.OS
}

func (m mockTransferStream) Send(res *pb.InstallResponse) error {
	switch res.Response.(type) {
	case *pb.InstallResponse_Validated:
		m.result <- res
	case *pb.InstallResponse_InstallError:
		m.result <- res
	default:
		m.response <- res
	}
	return nil
}

func (m mockTransferStream) Recv() (*pb.InstallRequest, error) {
	if request := m.errorReq; request != nil {
		return request, nil
	}
	select {
	case res := <-m.response:
		switch res.Response.(type) {
		case *pb.InstallResponse_TransferProgress:
			return &pb.InstallRequest{Request: &pb.InstallRequest_TransferEnd{}}, nil
		case *pb.InstallResponse_TransferReady:
			var out []byte
			if m.os.MockOS.Padding != nil {
				out, _ = proto.Marshal(&m.os.MockOS)
			} else {
				out = make([]byte, 10000000)
				rand.Read(out)
			}
			return &pb.InstallRequest{Request: &pb.InstallRequest_TransferContent{TransferContent: out}}, nil
		}
	default:
		return &pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{TransferRequest: &pb.TransferRequest{Version: m.os.MockOS.Version}}}, nil
	}
	return nil, nil
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
	buf := make([]byte, 10000000)
	rand.Read(buf)
	oS := &mockos.OS{MockOS: mockosPb.MockOS{
		Version: "1.0.2a",
		Cookie:  "cookiestring",
		Padding: buf,
	}}
	oS.Hash()
	tests := []struct {
		stream *mockTransferStream
		err    error
	}{
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				os:       oS,
			},
			err: nil,
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				errorReq: &pb.InstallRequest{Request: &pb.InstallRequest_TransferRequest{}}, // Unexpected request after transfer begins.
				os:       oS,
			},
			err: errors.New("Unknown request type"),
		},
	}
	for _, test := range tests {
		test.stream.response <- &pb.InstallResponse{Response: &pb.InstallResponse_TransferReady{}}
		_, err := ReceiveOS(test.stream)
		if diff := pretty.Compare(test.err, err); diff != "" {
			t.Errorf("ReceiveOS(stream): (-want +got):\n%s", diff)
		}
	}
}

func TestTargetInstall(t *testing.T) {
	buf := make([]byte, 10000000)
	rand.Read(buf)
	oS := &mockos.OS{MockOS: mockosPb.MockOS{
		Version: "1.0.2a",
		Cookie:  "cookiestring",
		Padding: buf,
	}}
	oS.Hash()
	incompatibleOS := &mockos.OS{MockOS: mockosPb.MockOS{
		Version:      "1.0.2b",
		Cookie:       "cookiestring",
		Padding:      buf,
		Incompatible: true,
	}}
	incompatibleOS.Hash()
	tests := []struct {
		stream *mockTransferStream
		want   *installResult
	}{
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os:       oS,
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_Validated{Validated: &pb.Validated{Version: oS.Version}}},
				err: nil,
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				errorReq: &pb.InstallRequest{Request: nil}, // Unexpected request.
				os:       oS,
			},
			want: &installResult{
				res: nil,
				err: errors.New("Failed to receive TransferRequest"),
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os: &mockos.OS{MockOS: mockosPb.MockOS{
					Version: "1.0.0a",
				}},
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{
					InstallError: &pb.InstallError{Type: pb.InstallError_INSTALL_RUN_PACKAGE},
				}},
				err: errors.New("Attempting to force transfer an OS of the same version as the currently running OS"),
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os: &mockos.OS{MockOS: mockosPb.MockOS{
					Version: "1.0.3c",
				}},
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_Validated{Validated: &pb.Validated{Version: "1.0.3c"}}},
				err: nil,
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os: &mockos.OS{MockOS: mockosPb.MockOS{
					Version: "1.0.2b",
					Cookie:  "cookiestring",
					Padding: buf,
					Hash:    []byte("BADHASH"),
				}},
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INTEGRITY_FAIL}}},
				err: nil,
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os:       incompatibleOS,
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INCOMPATIBLE, Detail: "Unsupported OS Version"}}},
				err: nil,
			},
		},
		{
			stream: &mockTransferStream{
				response: make(chan *pb.InstallResponse, 1),
				result:   make(chan *pb.InstallResponse, 1),
				os: &mockos.OS{MockOS: mockosPb.MockOS{
					Version: "1.0.2c",
				}},
			},
			want: &installResult{
				res: &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_PARSE_FAIL}}},
				err: nil,
			},
		},
	}
	for _, test := range tests {
		server := NewServer(&Settings{FactoryVersion: "1.0.0a", InstalledVersions: []string{"1.0.3c"}})
		got := &installResult{
			err: server.Install(test.stream),
		}
		close(test.stream.result)
		got.res = <-test.stream.result
		if diff := pretty.Compare(test.want, got); diff != "" {
			t.Errorf("Install(stream pb.OS_InstallServer): (-want +got):\n%s", diff)
		}
	}
}

func TestMultipleInstalls(t *testing.T) {
	buf := make([]byte, 10000000)
	rand.Read(buf)
	oS := &mockos.OS{MockOS: mockosPb.MockOS{
		Version: "1.0.2a",
		Cookie:  "cookiestring",
		Padding: buf,
	}}
	oS.Hash()
	server := NewServer(&Settings{FactoryVersion: "1.0.0a"})
	s1 := &mockTransferStream{
		response: make(chan *pb.InstallResponse, 1),
		result:   make(chan *pb.InstallResponse, 1),
		os:       oS,
	}
	s2 := &mockTransferStream{
		response: make(chan *pb.InstallResponse, 1),
		result:   make(chan *pb.InstallResponse, 1),
		os:       oS,
	}
	go server.Install(s1)
	go server.Install(s2)
	s1res := <-s1.result
	s2res := <-s2.result
	expect := &pb.InstallResponse{Response: &pb.InstallResponse_InstallError{InstallError: &pb.InstallError{Type: pb.InstallError_INSTALL_IN_PROGRESS}}}
	diff1 := pretty.Compare(expect, s1res)
	diff2 := pretty.Compare(expect, s2res)
	if (diff1 != "" && diff2 != "") || diff1 == diff2 {
		t.Errorf("Install(stream pb.OS_InstallServer): (-want +got):\n%s\n%s", diff1, diff2)
	}
}
