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
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/gnxi/gnoi/reset/pb"
)

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
		if got == nil && test.want == nil {
			continue
		} else {
			diff := cmp.Diff(test.want, got)
			log.Println(diff)
			if diff != "" {
				t.Errorf("CheckResponse(%s): (-want +got):\n%s", test.response, diff)
			}

		}
	}
}
