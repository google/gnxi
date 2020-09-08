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
	"path"
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
		targets  map[string]config.Target
		respBody string
	}{
		{
			"target not found",
			"name",
			http.StatusBadRequest,
			map[string]config.Target{},
			http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			"target found",
			"name",
			http.StatusOK,
			map[string]config.Target{"name": {}},
			"{\"address\":\"\",\"ca\":\"\",\"cakey\":\"\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("targets.devices", test.targets)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/target/%s", test.testName), nil)
			router := mux.NewRouter()
			router.HandleFunc("/target/{name}", handleTargetGet).Methods("GET")
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
		targets  map[string]config.Target
	}{
		{
			"empty body",
			"name",
			http.StatusInternalServerError,
			"\n",
			map[string]config.Target{},
		},
		{
			"setting target",
			"name",
			http.StatusOK,
			"{\"address\":\"test\",\"ca\":\"test.crt\",\"cakey\":\"test.key\"}\n",
			map[string]config.Target{"name": {
				Address: "test",
				Ca:      path.Join(filesDir(), "test.crt"),
				CaKey:   path.Join(filesDir(), "test.key"),
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("targets.devices", test.targets)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/target/%s", test.testName), bytes.NewBufferString(test.body))
			router := mux.NewRouter()
			router.HandleFunc("/target/{name}", handleTargetSet).Methods("POST")
			router.ServeHTTP(rr, req)
			if code := rr.Code; code != test.code {
				t.Errorf("Wanted exit code %d but got %d.", test.code, code)
			} else if diff := cmp.Diff(test.targets, viper.Get("targets.devices")); diff != "" {
				t.Errorf("Error in setting target (-want +got): %s", diff)
			}
		})
	}
}
