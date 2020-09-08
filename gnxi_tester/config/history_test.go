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
	"errors"
	"os"
	"path"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/viper"
)

type testCase struct {
	name          string
	targetName    string
	targetAddress string
	targetCA      string
	targetCAKey   string
	config        map[string]Target
	want          result
}

type result struct {
	targets map[string]Target
	err     error
}

func TestSetTarget(t *testing.T) {
	viper.SetConfigFile("/tmp/config.yml")
	tests := generateTargetTestCases()
	for _, test := range tests {
		viper.Reset()
		t.Run(test.name, func(t *testing.T) {
			var targets map[string]Target
			viper.Set("targets.devices", test.config)
			err := SetTarget(test.targetName, test.targetAddress, test.targetCA, test.targetCAKey, true)
			viper.UnmarshalKey("targets.devices", &targets)
			got := result{
				targets: targets,
				err:     err,
			}
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("prepareTarget(%s, %s, %s, %s): (-want +got)\n%s", test.targetName, test.targetAddress, test.targetCA, test.targetCAKey, diff)
			}
		})
	}
}

func TestPrepareTarget(t *testing.T) {
	tests := generateTargetTestCases()
	for _, test := range tests {
		viper.Reset()
		t.Run(test.name, func(t *testing.T) {
			var targets map[string]Target
			viper.Set("targets.devices", test.config)
			err := prepareTarget(test.targetName, test.targetAddress, test.targetCA, test.targetCAKey, true)
			viper.UnmarshalKey("targets.devices", &targets)
			got := result{
				targets: targets,
				err:     err,
			}
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("prepareTarget(%s, %s, %s, %s): (-want +got)\n%s", test.targetName, test.targetAddress, test.targetCA, test.targetCAKey, diff)
			}
		})
	}
}

func generateTargetTestCases() []testCase {
	dir, _ := os.Getwd()
	certPath := path.Join(dir, "ca.crt")
	certKeyPath := path.Join(dir, "ca.key")
	history := map[string]Target{"myhost.com": {
		Address: "localhost:9339",
		Ca:      certPath,
		CaKey:   certKeyPath,
	}}
	tests := []testCase{
		{
			name: "No targets in history, no target specified",
			want: result{
				err:     errors.New("No targets in history and no target specified"),
				targets: map[string]Target{},
			},
		},
		{
			name:   "No target specified",
			config: history,
			want: result{
				targets: history,
			},
		},
		{
			name:       "Non-existent target",
			config:     map[string]Target{},
			targetName: "nonexistenttarget",
			want: result{
				err:     errors.New("Target not found"),
				targets: history,
			},
		},
		{
			name:          "Add new target",
			targetName:    "myhost.com",
			targetAddress: "localhost:9339",
			targetCA:      certPath,
			targetCAKey:   certKeyPath,
			want: result{
				targets: history,
			},
		},
		{
			name:          "Update existing target's address",
			targetName:    "myhost.com",
			targetAddress: "newhost:9340",
			config:        history,
			want: result{
				targets: map[string]Target{
					"myhost.com": {
						Address: "newhost:9340",
						Ca:      certPath,
						CaKey:   certKeyPath,
					},
				},
			},
		},
		{
			name:       "Update existing target's ca",
			targetName: "myhost.com",
			targetCA:   "newca.crt",
			config:     history,
			want: result{
				targets: map[string]Target{
					"myhost.com": {
						Address: "localhost:9339",
						Ca:      path.Join(dir, "newca.crt"),
						CaKey:   certKeyPath,
					},
				},
			},
		},
		{
			name:        "Update existing target's ca_key",
			targetName:  "myhost.com",
			targetCAKey: "newca.key",
			config:      history,
			want: result{
				targets: map[string]Target{
					"myhost.com": {
						Address: "localhost:9339",
						Ca:      certPath,
						CaKey:   path.Join(dir, "newca.key"),
					},
				},
			},
		},
	}
	return tests
}
