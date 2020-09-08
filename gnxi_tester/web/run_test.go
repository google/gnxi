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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/spf13/viper"
)

func TestHandleRun(t *testing.T) {
	runTests = func(prompts config.Prompts, request runRequest) {}
	logErr = func(head http.Header, err error) {}
	tests := []struct {
		name    string
		prompts map[string]interface{}
		targets map[string]interface{}
		code    int
		postBody,
		respBody string
	}{
		{
			"failed to decode",
			map[string]interface{}{},
			map[string]interface{}{},
			http.StatusBadRequest,
			"not valid",
			http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			"prompts not found",
			map[string]interface{}{},
			map[string]interface{}{},
			http.StatusBadRequest,
			"{\"prompts\":\"name\"}",
			http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			"target not found",
			map[string]interface{}{"name": config.Prompts{}},
			map[string]interface{}{},
			http.StatusBadRequest,
			"{\"prompts\":\"name\"}",
			http.StatusText(http.StatusBadRequest) + "\n",
		},
		{
			"prompts found",
			map[string]interface{}{"name": config.Prompts{}},
			map[string]interface{}{"name": config.Target{}},
			http.StatusOK,
			"{\"prompts\":\"name\",\"target\":\"name\"}",
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("web.prompts", test.prompts)
			viper.Set("targets.devices", test.targets)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/run", bytes.NewBufferString(test.postBody))
			handler := http.HandlerFunc(handleRun)
			handler.ServeHTTP(rr, req)
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

func TestHandleRunOutput(t *testing.T) {
	test := struct {
		code int
		body string
	}{
		200,
		"test",
	}
	outputBuffer.WriteString(test.body)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/run/output", nil)
	handler := http.HandlerFunc(handleRunOutput)
	handler.ServeHTTP(rr, req)
	if code := rr.Code; code != test.code {
		t.Errorf("Wanted exit code %d but got %d.", test.code, code)
	}
	if b, err := ioutil.ReadAll(rr.Body); err != nil {
		t.Errorf("Error when decoding body: %w", err)
	} else if test.body != string(b) {
		t.Errorf("Wanted %s but got %s.", test.body, string(b))
	}
}
