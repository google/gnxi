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

func TestHandleTargetsGet(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody string
		targets  map[string]config.Target
	}{
		{
			name:     "Testing getting all targets, 1 target stored",
			wantCode: http.StatusOK,
			wantBody: `{"myhost.com":{"address":"localhost:9339","ca":"/ca.crt","cakey":"/ca.key"}}`,
			targets: map[string]config.Target{
				"myhost.com": {
					Address: "localhost:9339",
					Ca:      "/ca.crt",
					CaKey:   "/ca.key",
				},
			},
		},
		{
			name:     "Testing getting all targets, multiple targets stored",
			wantCode: http.StatusOK,
			wantBody: `{"anotherhost.com":{"address":"anotherlocalhost:9339","ca":"/anotherca.crt","cakey":"/anotherca.key"},"myhost.com":{"address":"localhost:9339","ca":"/ca.crt","cakey":"/ca.key"}}`,
			targets: map[string]config.Target{
				"myhost.com": {
					Address: "localhost:9339",
					Ca:      "/ca.crt",
					CaKey:   "/ca.key",
				},
				"anotherhost.com": {
					Address: "anotherlocalhost:9339",
					Ca:      "/anotherca.crt",
					CaKey:   "/anotherca.key",
				},
			},
		},
		{
			name:     "Testing getting all targets, no targets stored",
			wantCode: http.StatusOK,
			wantBody: `{}`,
			targets:  map[string]config.Target{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("targets.devices", test.targets)
			resRecorder := httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/target", nil)
			router := mux.NewRouter()
			router.HandleFunc("/target", handleTargetsGet).Methods("GET")
			router.ServeHTTP(resRecorder, request)
			if code := resRecorder.Code; code != test.wantCode {
				t.Errorf("Expected code %d, got %d", test.wantCode, code)
			} else if diff := cmp.Diff(test.wantBody, string(resRecorder.Body.Bytes())); diff != "" {
				t.Errorf("Error in getting all targets (-want +got): %s", diff)
			}
		})
	}
}

func TestHandleTargetDelete(t *testing.T) {
	tests := []struct {
		name        string
		targetName  string
		wantTargets map[string]config.Target
		targets     map[string]config.Target
	}{
		{
			name:        "Deleting 1 target, 1 target stored",
			targetName:  "myhost.com",
			wantTargets: map[string]config.Target{},
			targets: map[string]config.Target{
				"myhost.com": {
					Address: "localhost:9339",
					Ca:      "/ca.crt",
					CaKey:   "/ca.key",
				},
			},
		},
		{
			name:       "Deleting 1 target, multiple targets stored",
			targetName: "anotherhost.com",
			wantTargets: map[string]config.Target{
				"myhost.com": {
					Address: "localhost:9339",
					Ca:      "/ca.crt",
					CaKey:   "/ca.key",
				},
			},
			targets: map[string]config.Target{
				"myhost.com": {
					Address: "localhost:9339",
					Ca:      "/ca.crt",
					CaKey:   "/ca.key",
				},
				"anotherhost.com": {
					Address: "anotherlocalhost:9339",
					Ca:      "/anotherca.crt",
					CaKey:   "/anotherca.key",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("targets.devices", test.targets)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/target/%s", test.targetName), nil)
			router := mux.NewRouter()
			router.HandleFunc("/target/{name}", handleTargetDelete).Methods("DELETE")
			router.ServeHTTP(rr, req)
			if diff := cmp.Diff(test.wantTargets, viper.Get("targets.devices")); diff != "" {
				t.Errorf("Error in deleting target (-want +got): %s", diff)
			}
		})
	}
}
