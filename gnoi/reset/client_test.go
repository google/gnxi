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
			response: initializeResponse(true, false, false, ""),
			want:     &ResetError{[]string{"Factory OS Rollback Unsupported"}},
		},
		{
			response: initializeResponse(false, true, false, ""),
			want:     &ResetError{[]string{"Zero Filling Persistent Storage Unsupported"}},
		},
		{
			response: initializeResponse(false, false, true, "Unspecified Test Error"),
			want:     &ResetError{[]string{"Unspecified Error: Unspecified Test Error"}},
		},
		{
			response: initializeResponse(true, true, false, ""),
			want: &ResetError{[]string{
				"Factory OS Rollback Unsupported",
				"Zero Filling Persistent Storage Unsupported",
			}},
		},
		{
			response: initializeResponse(true, true, true, "Unspecified Test Error"),
			want: &ResetError{[]string{
				"Factory OS Rollback Unsupported",
				"Zero Filling Persistent Storage Unsupported",
				"Unspecified Error: Unspecified Test Error",
			}},
		},
	}
	for _, test := range tests {
		got := CheckResponse(test.response)
		diff := cmp.Diff(test.want, got)
		if diff != "" {
			t.Errorf("CheckResponse(%s): (-want +got):\n%s", test.response, diff)
		}
	}
}

func initializeResponse(factoryOSUnsupported, zeroFillUnsupported, other bool, details string) *pb.StartResponse {
	res := &pb.StartResponse{}
	if factoryOSUnsupported || zeroFillUnsupported || other {
		resetError := &pb.ResetError{
			FactoryOsUnsupported: factoryOSUnsupported,
			ZeroFillUnsupported:  zeroFillUnsupported,
			Other:                other,
			Detail:               details,
		}
		res.Response = &pb.StartResponse_ResetError{resetError}
	} else {
		resetSuccess := &pb.ResetSuccess{}
		res.Response = &pb.StartResponse_ResetSuccess{resetSuccess}
	}
	return res
}
