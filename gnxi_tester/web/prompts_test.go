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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestHandlePromptsSet(t *testing.T) {
	logErr = func(head http.Header, err error) {}
	tests := []struct {
		name    string
		code    int
		prompts *config.Prompts
		reqBody string
	}{
		{
			"decode error",
			http.StatusBadRequest,
			nil,
			"",
		},
		{
			"terminates correctly",
			http.StatusOK,
			&config.Prompts{
				Name: "test",
				Prompts: map[string]string{
					"test": "test",
				},
				Files: map[string]string{
					"test": "test",
				},
			},
			"{\"name\":\"test\",\"prompts\":{\"test\":\"test\"},\"files\":{\"test\":\"test\"}}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("web.prompts", map[string]config.Prompts{})
			viper.SetConfigFile("/tmp/config.yml")
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/prompts", strings.NewReader(test.reqBody))
			handler := http.HandlerFunc(handlePromptsSet)
			handler.ServeHTTP(rr, req)
			got := viper.GetStringMap("web.prompts")
			if code := rr.Code; code != test.code {
				t.Errorf("Wanted exit code %d but got %d.", test.code, code)
			}
			if test.prompts != nil {
				if diff := cmp.Diff(*test.prompts, got[test.prompts.Name]); diff != "" {
					t.Errorf("Prompt incorrectly set (-want +got): %s.", diff)
				}
			}
		})
	}
}

func TestHandlePromptsGet(t *testing.T) {
	logErr = func(head http.Header, err error) {}
	tests := []struct {
		name     string
		code     int
		prompts  map[string]interface{}
		respBody string
	}{
		{
			"encode error",
			http.StatusOK,
			nil,
			"[]\n",
		},
		{
			"terminates correctly",
			http.StatusOK,
			map[string]interface{}{
				"test_name": config.Prompts{
					Name: "test_name",
					Prompts: map[string]string{
						"test_prompt": "test_value",
					},
					Files: map[string]string{
						"test_file": "test_path",
					},
				},
			},
			"[{\"name\":\"test_name\",\"prompts\":{\"test_prompt\":\"test_value\"},\"files\":{\"test_file\":\"test_path\"}}]\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.SetConfigFile("/tmp/config.yml")
			viper.Set("web.prompts", test.prompts)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/prompts", nil)
			handler := http.HandlerFunc(handlePromptsGet)
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
