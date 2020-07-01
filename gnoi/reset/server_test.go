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
	"fmt"
	"testing"

	"github.com/google/gnxi/gnoi/reset/pb"
)

type test struct {
	name     string
	request  *pb.StartRequest
	settings Settings
}

func makeTests() []test {
	reqs := []pb.StartRequest{
		{ZeroFill: true},
		{ZeroFill: true, FactoryOs: true},
		{FactoryOs: true},
		{},
	}
	sets := []Settings{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}
	t := []test{}
	for _, set := range sets {
		for _, req := range reqs {
			name := fmt.Sprintf(
				"ZF:{Given:%v, Unsupported:%v},OS:{Given:%v,Unsupported:%v}",
				req.ZeroFill,
				set.ZeroFillUnsupported,
				req.FactoryOs,
				set.FactoryOSUnsupported,
			)
			t = append(t, test{name: name, request: &req, settings: set})
		}
	}
	return t
}

func TestStart(t *testing.T) {
	tests := makeTests()
	for _, test := range tests {
		s := NewServer(&test.settings)
		t.Run(test.name, func(t *testing.T) {
			resp, err := s.Start(context.Background(), test.request)
			if err != nil {
				t.Fatalf("ops")
			}
			switch response := resp.Response.(type) {
			case *pb.StartResponse_ResetSuccess:
				if test.settings.ZeroFillUnsupported && test.request.ZeroFill || test.settings.FactoryOSUnsupported && test.request.FactoryOs {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseSuccess{%v},",
						test.settings.ZeroFillUnsupported,
						test.settings.FactoryOSUnsupported,
						test.request.ZeroFill,
						test.request.FactoryOs,
						response)
				}
			case *pb.StartResponse_ResetError:
				if response.ResetError.ZeroFillUnsupported != (test.settings.ZeroFillUnsupported && test.request.ZeroFill) ||
					response.ResetError.FactoryOsUnsupported != (test.settings.FactoryOSUnsupported && test.request.FactoryOs) {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseError{%v},",
						test.settings.ZeroFillUnsupported,
						test.settings.FactoryOSUnsupported,
						test.request.ZeroFill,
						test.request.FactoryOs,
						response)
				}
			}
		})
	}
}
