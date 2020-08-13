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

package main

import (
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/proto/gnmi"
)

type MockClientStream struct {
	gnmi.GNMI_SubscribeClient
	responses   chan *gnmi.SubscribeResponse
	pollingMode bool
}

func (m MockClientStream) Send(req *gnmi.SubscribeRequest) error {
	return nil
}

func (m MockClientStream) Recv() (*gnmi.SubscribeResponse, error) {
	select {
	case res := <-m.responses:
		return res, nil
	default:
		return nil, io.EOF
	}
}

func TestOnce(t *testing.T) {
	tests := []struct {
		name      string
		responses []*gnmi.SubscribeResponse
		want      error
	}{
		{
			name: "Send 2 Updates then SyncResponse",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 0}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 1}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
		},
		{
			name: "Send SyncResponse",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
		},
		{
			name: "Send Unknown Response",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Error{}},
			},
			want: errors.New("unexpected response type"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := &MockClientStream{
				responses: make(chan *gnmi.SubscribeResponse, len(test.responses)),
			}
			for _, response := range test.responses {
				stream.responses <- response
			}
			got := once(stream)
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("once(): (-want +got)\n%s", diff)
			}
		})
	}
}

func TestStream(t *testing.T) {
	tests := []struct {
		name      string
		responses []*gnmi.SubscribeResponse
		want      error
	}{
		{
			name: "Send Unknown Response",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Error{}},
			},
			want: errors.New("unexpected response type"),
		},
		{
			name: "Send a 3 requests followed by a sync response 3 times",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 0}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 1}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 2}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 3}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 4}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 5}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 6}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 7}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 8}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientStream := &MockClientStream{
				responses: make(chan *gnmi.SubscribeResponse, len(test.responses)),
			}
			for _, response := range test.responses {
				clientStream.responses <- response
			}
			got := stream(clientStream)
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("stream(): (-want +got)\n%s", diff)
			}
		})
	}
}

func TestPoll(t *testing.T) {
	tests := []struct {
		name        string
		responses   []*gnmi.SubscribeResponse
		updatesOnly bool
		want        error
	}{
		{
			name: "Send Unknown Response",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Error{}},
			},
			want: errors.New("unexpected response type"),
		},
		{
			name: "Send a 3 requests followed by a SyncResponse 3 times",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 0}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 1}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 2}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 3}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 4}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 5}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 6}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 7}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 8}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
		},
		{
			name: "Test -updatesOnly flag",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 0}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 1}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 2}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 3}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 4}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 5}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 6}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 7}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 8}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
			updatesOnly: true,
		},
		{
			name: "Test -updatesOnly flag without sending initial SyncResponse",
			responses: []*gnmi.SubscribeResponse{
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 0}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 1}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 2}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 3}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 4}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 5}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 6}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 7}}},
				{Response: &gnmi.SubscribeResponse_Update{Update: &gnmi.Notification{Timestamp: 8}}},
				{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}},
			},
			updatesOnly: true,
			want:        errors.New("-updates_only flag is set but failed to receive SyncResponse first for POLL mode"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientStream := &MockClientStream{
				responses: make(chan *gnmi.SubscribeResponse, len(test.responses)),
			}
			for _, response := range test.responses {
				clientStream.responses <- response
			}
			got := poll(clientStream, test.updatesOnly, testPollInput)
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("poll(): (-want +got)\n%s", diff)
			}
		})
	}
}

func testPollInput() {
	return
}

func TestAssembleSubscriptions(t *testing.T) {
	defaultPath := &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "system"}}}
	tests := []struct {
		name           string
		paths          []*gnmi.Path
		streamOnChange bool
		sampleInterval uint64
	}{
		{
			name:           "Flags -stream_on_change and -sample_interval set",
			paths:          []*gnmi.Path{defaultPath},
			streamOnChange: true,
			sampleInterval: 1,
		},
		{
			name:  "No SubscriptionMode flags set (TARGET_DEFINED)",
			paths: []*gnmi.Path{defaultPath},
		},
		{
			name:           "-sample_interval set to 5",
			paths:          []*gnmi.Path{defaultPath},
			sampleInterval: 5,
		},
		{
			name:           "-stream_on_change set to true",
			paths:          []*gnmi.Path{defaultPath},
			streamOnChange: true,
		},
	}
	wants := []struct {
		err           error
		subscriptions []*gnmi.Subscription
	}{
		{err: errors.New("only one of -stream_on_change and -sample_interval can be set")},
		{
			subscriptions: []*gnmi.Subscription{
				{
					Path: defaultPath,
					Mode: gnmi.SubscriptionMode_TARGET_DEFINED,
				},
			},
		},
		{
			subscriptions: []*gnmi.Subscription{
				{
					Path:           defaultPath,
					Mode:           gnmi.SubscriptionMode_SAMPLE,
					SampleInterval: 5,
				},
			},
		},
		{
			subscriptions: []*gnmi.Subscription{
				{
					Path: defaultPath,
					Mode: gnmi.SubscriptionMode_ON_CHANGE,
				},
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			subscriptionsGot, errGot := assembleSubscriptions(test.streamOnChange, test.sampleInterval, test.paths)
			got := struct {
				err           error
				subscriptions []*gnmi.Subscription
			}{
				err:           errGot,
				subscriptions: subscriptionsGot,
			}
			if diff := pretty.Compare(wants[i], got); diff != "" {
				t.Errorf("assembleSubscriptions(%v, %d, %v): (-want +got)\n%s", test.streamOnChange, test.sampleInterval, test.paths, diff)
			}
		})
	}
}

func TestParseEncoding(t *testing.T) {
	tests := []struct {
		encoding     string
		wantEncoding gnmi.Encoding
		wantErr      string
	}{
		{
			encoding:     "JSON",
			wantEncoding: gnmi.Encoding_JSON,
		},
		{
			encoding:     "JSON_IETF",
			wantEncoding: gnmi.Encoding_JSON_IETF,
		},
		{
			encoding:     "PROTO",
			wantEncoding: gnmi.Encoding_PROTO,
		},
		{
			encoding:     "ASCII",
			wantEncoding: gnmi.Encoding_ASCII,
		},
		{
			encoding:     "BYTES",
			wantEncoding: gnmi.Encoding_BYTES,
		},
		{
			encoding: "NON_EXISTANT_FORMAT",
			wantErr:  "supported encodings:",
		},
	}
	for _, test := range tests {
		got, gotErr := parseEncoding(test.encoding)
		if gotErr != nil && !strings.Contains(gotErr.Error(), test.wantErr) {
			t.Errorf("Expected error to contain %s: Got %v", test.wantErr, gotErr)
		}
		if got != test.wantEncoding {
			t.Errorf("Got %s, want %s", gnmi.Encoding_name[int32(got)], test.encoding)
		}
	}
}

func TestSubscriptionMode(t *testing.T) {
	tests := []struct {
		poll bool
		once bool
	}{
		{},
		{poll: true},
		{once: true},
		{
			poll: true,
			once: true,
		},
	}
	wants := []struct {
		mode gnmi.SubscriptionList_Mode
		err  error
	}{
		{mode: gnmi.SubscriptionList_STREAM},
		{mode: gnmi.SubscriptionList_POLL},
		{mode: gnmi.SubscriptionList_ONCE},
		{err: errors.New("only one of -once and -poll can be set")},
	}
	for i, test := range tests {
		modeGot, errGot := subscriptionMode(test.poll, test.once)
		got := struct {
			mode gnmi.SubscriptionList_Mode
			err  error
		}{
			mode: modeGot,
			err:  errGot,
		}
		if diff := pretty.Compare(wants[i], got); diff != "" {
			t.Errorf("subscriptionMode(%v, %v): (-want +got)\n%s", test.poll, test.once, diff)
		}
	}
}

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name        string
		xPathFlags  arrayFlags
		pbPathFlags arrayFlags
	}{
		{
			name:       "XPath without keys",
			xPathFlags: arrayFlags{"/system"},
		},
		{
			name:       "XPath with keys",
			xPathFlags: arrayFlags{"/controller[name=main]"},
		},
		{
			name:       "Invalid XPath",
			xPathFlags: arrayFlags{`\`},
		},
	}
	wants := []struct {
		paths []*gnmi.Path
		err   string
	}{
		{paths: []*gnmi.Path{{Elem: []*gnmi.PathElem{{Name: "system"}}}}},
		{
			paths: []*gnmi.Path{{Elem: []*gnmi.PathElem{{
				Name: "controller",
				Key:  map[string]string{"name": "main"},
			}}}},
		},
		{err: `error in parsing xpath`},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths, err := parsePaths(test.xPathFlags, test.pbPathFlags)
			if err != nil {
				if match, _ := regexp.Match(wants[i].err, []byte(err.Error())); !match {
					t.Errorf("Got error %v, did not match %s", err, wants[i].err)
				}
			}
			if diff := pretty.Compare(wants[i].paths, paths); diff != "" {
				t.Errorf("parsePaths(%v, %v): (-want +got)\n%s", test.xPathFlags, test.pbPathFlags, diff)
			}
		})
	}
}
