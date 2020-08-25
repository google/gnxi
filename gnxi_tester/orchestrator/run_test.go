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

package orchestrator

import (
	"errors"
	"testing"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestRunTests(t *testing.T) {
	tests := []struct {
		name         string
		testNames    []string
		tests        map[string][]config.Test
		order        []string
		files        map[string][]string
		prompt       callbackFunc
		wantSucc     []string
		wantErr      error
		runContainer func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error)
	}{
		{
			"Run all tests",
			[]string{},
			map[string][]config.Test{
				"test": {{Name: "test"}, {Name: "test2"}},
			},
			[]string{"test"},
			map[string][]string{"test": {"tt"}},
			func(name string) string { return name },
			[]string{"*test*:\ntest:\ntest\n\ntest2:\ntest\n"},
			nil,
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = name
				return
			},
		},
		{
			"Run all tests with prompt",
			[]string{},
			map[string][]config.Test{
				"test": {{Name: "test", Args: map[string]string{"ask": "&<ask>"}, Prompt: []string{"ask"}}},
			},
			[]string{"test"},
			map[string][]string{},
			func(name string) string { return name },
			[]string{"*test*:\ntest:\n-ask ask -logtostderr -target_name test -target_addr test -ca /certs/ca.crt -ca_key /certs/ca.key\n"},
			nil,
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = args
				return
			},
		},
		{
			"Run one test",
			[]string{"test"},
			map[string][]config.Test{
				"test":  {{Name: "test"}},
				"test2": {{Name: "test2"}},
			},
			[]string{"test", "test2"},
			map[string][]string{"test": {"tt"}},
			func(name string) string { return name },
			[]string{"*test*:\ntest:\ntest\n"},
			nil,
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = name
				return
			},
		},
		{
			"Tests want correct",
			[]string{},
			map[string][]config.Test{
				"test": {{Name: "test", Wants: "test"}},
			},
			[]string{"test"},
			map[string][]string{},
			func(name string) string { return name },
			[]string{"*test*:\ntest:\ntest\n"},
			nil,
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = name
				return
			},
		},
		{
			"Tests want incorrect",
			[]string{},
			map[string][]config.Test{
				"test": {{Name: "test", Wants: "no"}},
			},
			[]string{"test"},
			map[string][]string{},
			func(name string) string { return name },
			nil,
			formatErr("test", "test", "test", errors.New("Wanted no in output"), 0, false, "-logtostderr -target_name test -target_addr test -ca /certs/ca.crt -ca_key /certs/ca.key", nil),
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = name
				return
			},
		},
		{
			"Tests don't want correct",
			[]string{},
			map[string][]config.Test{
				"test": {{Name: "test", DoesntWant: "aaaa"}},
			},
			[]string{"test"},
			map[string][]string{},
			func(name string) string { return name },
			[]string{"*test*:\ntest:\ntest\n"},
			nil,
			func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
				out = name
				return
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			InitContainers = func(names []string) error { return nil }
			viper.Set("targets.devices", map[string]config.Device{"test": {Address: "test", Ca: "/certs/ca.crt", CaKey: "/certs/ca.key"}})
			viper.Set("targets.last_target", "test")
			viper.Set("tests", test.tests)
			viper.Set("order", test.order)
			viper.Set("files", test.files)
			RunContainer = test.runContainer
			succ, err := RunTests(test.testNames, test.prompt, map[string]string{}, func(string, ...interface{}) {})
			if diff := cmp.Diff(test.wantSucc, succ); diff != "" {
				t.Errorf("(-want +got): %s", diff)
			} else if (test.wantErr == nil) != (err == nil) {
				t.Errorf("invalid error: want: %v, got: %v", test.wantErr, err)
			}
		})
	}
}
