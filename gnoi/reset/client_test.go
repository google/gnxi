/* Copyright 2018 Google Inc.

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

	"github.com/google/gnxi/gnoi/reset/pb"
)

func TestStart(t *testing.T) {
	t.Run("ResetSuccess", func(t *testing.T) {
		res := initializeResponse(false, false, false, "")
		err := CheckResponse(res)
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
	})

	t.Run("ResetError OS Rollback Impossible", func(t *testing.T) {
		res := initializeResponse(true, false, false, "")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected OS unsupported error, no error returned")
		}
	})
	t.Run("ResetError Zero Fill Impossible", func(t *testing.T) {
		res := initializeResponse(false, true, false, "")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected Zero Fill unsupported error, no error returned")
		}
	})
	t.Run("Unspecified Error Response", func(t *testing.T) {
		res := initializeResponse(false, false, true, "Unspecified Test Error")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected Unspecified error, no error returned")
		}
	})
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
