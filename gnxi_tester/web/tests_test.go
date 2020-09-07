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

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func TestHandleTestsOrderGet(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody string
		order    []string
	}{
		{
			name:     "Testing getting set of test names",
			wantCode: http.StatusOK,
			wantBody: `["test1","test2","test3"]` + "\n",
			order:    []string{"test1", "test2", "test3"},
		},
		{
			name:     "Testing getting test order with no order specified",
			wantCode: http.StatusOK,
			wantBody: `[]` + "\n",
			order:    []string{},
		},
	}
	for _, test := range tests {
		viper.Reset()
		t.Run(test.name, func(t *testing.T) {
			viper.Set("order", test.order)
			resRecorder := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/test/order", nil)
			router := mux.NewRouter()
			router.HandleFunc("/test/order", handleTestsOrderGet).Methods("GET")
			router.ServeHTTP(resRecorder, request)
			if code := resRecorder.Code; code != test.wantCode {
				t.Errorf("Expected code %d, got %d", test.wantCode, code)
			} else if diff := cmp.Diff(test.wantBody, string(resRecorder.Body.Bytes())); diff != "" {
				t.Errorf("Error in getting test order (-want +got): %s", diff)
			}
		})
	}
}

func TestHandleTestsGet(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody string
		tests    map[string][]config.Test
	}{
		{
			name:     "Testing getting all tests, 1 test stored",
			wantCode: http.StatusOK,
			wantBody: `{"major_test":[{"name":"minor_test 1","args":{},"mustfail":false,"wait":0,"wants":"","doesntwant":"","prompt":[]}]}` + "\n",
			tests: map[string][]config.Test{
				"major_test": {
					{
						Name:   "minor_test 1",
						Args:   map[string]string{},
						Prompt: []string{},
					},
				},
			},
		},
		{
			name:     "Testing getting all tests, no tests stored",
			wantCode: http.StatusOK,
			wantBody: `{}` + "\n",
			tests:    map[string][]config.Test{},
		},
	}
	for _, test := range tests {
		viper.Reset()
		t.Run(test.name, func(t *testing.T) {
			viper.Set("tests", test.tests)
			resRecorder := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/test", nil)
			router := mux.NewRouter()
			router.HandleFunc("/test", handleTestsGet).Methods("GET")
			router.ServeHTTP(resRecorder, request)
			if code := resRecorder.Code; code != test.wantCode {
				t.Errorf("Expected code %d, got %d", test.wantCode, code)
			} else if diff := cmp.Diff(test.wantBody, string(resRecorder.Body.Bytes())); diff != "" {
				t.Errorf("Error in getting all tests (-want +got): %s", diff)
			}
		})
	}
}
