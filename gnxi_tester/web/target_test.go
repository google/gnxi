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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func TestHandleTargetGet(t *testing.T) {
	logErr = func(head http.Header, err error) {}
	tests := []struct {
		name     string
		testName string
		code     int
		devices  map[string]config.Device
		respBody string
	}{
		{
			"device not found",
			"name",
			http.StatusBadRequest,
			map[string]config.Device{},
			http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			"device found",
			"name",
			http.StatusOK,
			map[string]config.Device{"name": {}},
			"{\"address\":\"\",\"ca\":\"\",\"cakey\":\"\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("targets.devices", test.devices)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/device/%s", test.testName), nil)
			router := mux.NewRouter()
			router.HandleFunc("/device/{name}", handleTargetGet).Methods("GET")
			router.ServeHTTP(rr, req)
			if code := rr.Code; code != test.code {
				t.Errorf("Wanted exit code %d but got %d.", test.code, code)
			}
			if b, err := ioutil.ReadAll(rr.Body); err != nil {
				t.Errorf("Error when decoding body: %w", err)
			} else if test.respBody != string(b) {
				t.Errorf("Wanted %s but got %s.", test.respBody, string(b))
			}
		})
	}
}

func TestHandleTargetSet(t *testing.T) {
	logErr = func(head http.Header, err error) {}
	tests := []struct {
		name     string
		testName string
		code     int
		body     string
		devices  map[string]config.Device
	}{
		{
			"empty body",
			"name",
			http.StatusInternalServerError,
			"\n",
			map[string]config.Device{},
		},
		{
			"device found",
			"name",
			http.StatusOK,
			"{\"address\":\"\",\"ca\":\"\",\"cakey\":\"\"}\n",
			map[string]config.Device{"name": {}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("targets.devices", test.devices)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/device/%s", test.testName), bytes.NewBufferString(test.body))
			router := mux.NewRouter()
			router.HandleFunc("/device/{name}", handleTargetSet).Methods("POST")
			router.ServeHTTP(rr, req)
			if code := rr.Code; code != test.code {
				t.Errorf("Wanted exit code %d but got %d.", test.code, code)
			} else if diff := cmp.Diff(test.devices, viper.Get("targets.devices")); diff != "" {
				t.Errorf("Error in setting devicve (-want +got): %s", diff)
			}
		})
	}
}
