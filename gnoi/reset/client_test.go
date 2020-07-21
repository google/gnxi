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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc"

	"github.com/google/gnxi/gnoi/reset/pb"
)

type mockClient struct {
	server Settings
}

func (m *mockClient) Start(ctx context.Context, request *pb.StartRequest, opts ...grpc.CallOption) (*pb.StartResponse, error) {
	resetError := &pb.ResetError{}
	resetError.ZeroFillUnsupported = request.ZeroFill && m.server.ZeroFillUnsupported
	resetError.FactoryOsUnsupported = request.FactoryOs && m.server.FactoryOSUnsupported

	if resetError.ZeroFillUnsupported || resetError.FactoryOsUnsupported {
		return &pb.StartResponse{Response: &pb.StartResponse_ResetError{ResetError: resetError}}, nil
	}
	return &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{}}, nil
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		response *pb.StartResponse
		want     *ResetError
	}{
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetSuccess{&pb.ResetSuccess{}}},
			want:     nil,
		},
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetError{&pb.ResetError{
				FactoryOsUnsupported: true,
				ZeroFillUnsupported:  false,
				Other:                false,
				Detail:               "",
			}}},
			want: &ResetError{[]string{"Factory OS Rollback Unsupported"}},
		},
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetError{&pb.ResetError{
				FactoryOsUnsupported: false,
				ZeroFillUnsupported:  true,
				Other:                false,
				Detail:               "",
			}}},
			want: &ResetError{[]string{"Zero Filling Persistent Storage Unsupported"}},
		},
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetError{&pb.ResetError{
				FactoryOsUnsupported: false,
				ZeroFillUnsupported:  false,
				Other:                true,
				Detail:               "Unspecified Test Error",
			}}},
			want: &ResetError{[]string{"Unspecified Error: Unspecified Test Error"}},
		},
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetError{&pb.ResetError{
				FactoryOsUnsupported: true,
				ZeroFillUnsupported:  true,
				Other:                false,
				Detail:               "",
			}}},
			want: &ResetError{[]string{
				"Factory OS Rollback Unsupported",
				"Zero Filling Persistent Storage Unsupported",
			}},
		},
		{
			response: &pb.StartResponse{Response: &pb.StartResponse_ResetError{&pb.ResetError{
				FactoryOsUnsupported: true,
				ZeroFillUnsupported:  true,
				Other:                true,
				Detail:               "Unspecified Test Error",
			}}},
			want: &ResetError{[]string{
				"Factory OS Rollback Unsupported",
				"Zero Filling Persistent Storage Unsupported",
				"Unspecified Error: Unspecified Test Error",
			}},
		},
	}
	for _, test := range tests {
		got := CheckResponse(test.response)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("CheckResponse(%s): (-want +got):\n%s", test.response, diff)
		}
	}
}

func TestResetTarget(t *testing.T) {
	requests := []*pb.StartRequest{
		{},
		{FactoryOs: true},
		{ZeroFill: true},
		{
			FactoryOs: true,
			ZeroFill:  true,
		},
	}
	settings := []Settings{
		{},
		{FactoryOSUnsupported: true},
		{ZeroFillUnsupported: true},
		{
			FactoryOSUnsupported: true,
			ZeroFillUnsupported:  true,
		},
	}
	for _, setting := range settings {
		cli := &Client{client: &mockClient{server: setting}}
		for _, req := range requests {
			var expect *ResetError
			got := cli.ResetTarget(context.Background(), req.ZeroFill, req.FactoryOs)
			if (setting.FactoryOSUnsupported && req.FactoryOs) && (setting.ZeroFillUnsupported && req.ZeroFill) {
				expect = &ResetError{
					Msgs: []string{"Factory OS Rollback Unsupported", "Zero Filling Persistent Storage Unsupported"},
				}
			} else if setting.FactoryOSUnsupported && req.FactoryOs {
				expect = &ResetError{
					Msgs: []string{"Factory OS Rollback Unsupported"},
				}
			} else if setting.ZeroFillUnsupported && req.ZeroFill {
				expect = &ResetError{
					Msgs: []string{"Zero Filling Persistent Storage Unsupported"},
				}
			}
			if diff := cmp.Diff(expect, got); diff != "" {
				t.Errorf("ResetTarget(context.Background(), %v, %v): (-want +got)\n%s", req.ZeroFill, req.FactoryOs, diff)
			}
		}
	}
}

func TestResetError(t *testing.T) {
	tests := []struct {
		error *ResetError
		want  string
	}{
		{
			error: &ResetError{Msgs: []string{"Factory OS Rollback Unsupported"}},
			want:  "Factory OS Rollback Unsupported",
		},
		{
			error: &ResetError{Msgs: []string{"Factory OS Rollback Unsupported", "Zero Filling Persistent Storage Unsupported"}},
			want:  "Factory OS Rollback Unsupported\nZero Filling Persistent Storage Unsupported",
		},
	}
	for _, test := range tests {
		got := test.error.Error()
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("%v: (-want +got)\n%s", test.error, diff)
		}
	}
}

func TestNewClient(t *testing.T) {
	conn := &grpc.ClientConn{}
	want := &Client{client: pb.NewFactoryResetClient(conn)}
	got := NewClient(conn)
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("(NewClient(): (-want +got)\n%s", diff)
	}
}
