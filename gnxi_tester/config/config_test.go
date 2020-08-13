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

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestGetTests(t *testing.T) {
	wants := []Tests{
		{},
		{"gnoi_os": []Test{{
			Name:   "Compatible OS with Good Hash Install",
			Args:   map[string]string{"op": "install", "version": "&<version>", "os": "&<os_path>"},
			Wants:  `^$`,
			Prompt: []string{"version", "os_path"},
		}}},
		{
			"gnoi_cert": []Test{
				{
					Name:  "Rotate Certificate",
					Args:  map[string]string{"op": "rotate", "cert_id": "&<cert_id>"},
					Wants: "Rotate success",
				},
				{
					Name:     "Rotate Non-Existent Certificate",
					Args:     map[string]string{"op": "rotate", "cert_id": "&<non_existent_cert_id>"},
					MustFail: true,
					Wants:    "Failed Rotate",
					Prompt:   []string{"non_existent_cert_id"},
				},
			},
			"gnoi_reset": []Test{{
				Name:  "Resetting a Target Successfully",
				Args:  map[string]string{},
				Wants: `^$`,
			}},
		},
	}
	for _, want := range wants {
		viper.Set("tests", want)
		got := GetTests()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GetTests(): (-got +want):\n%s", diff)
		}
	}
}

func TestGetDevices(t *testing.T) {
	tests := []map[string]Device{
		{},
		{"mydevice.com": Device{
			Address: "localhost:9339",
			Ca:      "ca.crt",
			CaKey:   "ca.key",
		}},
		{
			"mydevice.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			},
			"anotherdevice.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			},
		},
	}
	for _, cfg := range tests {
		viper.Reset()
		viper.Set("targets.devices", cfg)
		got := GetDevices()
		if diff := cmp.Diff(cfg, got); diff != "" {
			t.Errorf("GetDevices(): (-want +got):\n%s", diff)
		}
	}
}
