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
	"github.com/spf13/viper"
)

// Tests represent a set of major tests.
type Tests map[string][]Test

// Test represents a single set of inputs and expected outputs.
type Test struct {
	Name       string            `mapstructure:"name"`
	Args       map[string]string `mapstructure:"args"`
	MustFail   bool              `mapstructure:"must_fail"`
	Wait       int               `mapstructure:"wait"`
	Wants      string            `mapstructure:"wants"`
	DoesntWant string            `mapstructure:"doesnt_want"`
	Prompt     []string          `mapstructure:"prompt"`
}

func setDefaults() {
	testCases := generateTestCases()
	viper.SetDefault("tests", testCases)
	viper.SetDefault("docker.build", "golang:1.14-alpine")
	viper.SetDefault("docker.runtime", "alpine")
	viper.SetDefault("docker.files", createDockerfiles(testCases))
}

func generateTestCases() Tests {
	certTests := []Test{
		{
			Name:   "Provision Bootstrapping Target",
			Args:   map[string]string{"op": "provision", "cert_id": "&<cert_id>"},
			Wants:  "Install success",
			Prompt: []string{"cert_id"},
		},
		{
			Name:  "Get certs",
			Args:  map[string]string{"op": "get"},
			Wait:  10,
			Wants: `GetCertificates:\n{.*}$`,
		},
		{
			Name:   "Install a cert",
			Args:   map[string]string{"op": "install", "cert_id": "&<new_cert_id>"},
			Wants:  "Install success",
			Prompt: []string{"new_cert_id"},
		},
		{
			Name:  "Check if Target Can Generate CSR",
			Args:  map[string]string{"op": "check"},
			Wants: `CanGenerateCSR:\ntrue`,
		},
		{
			Name:       "Revoke a cert",
			Args:       map[string]string{"op": "revoke", "cert_id": "&<new_cert_id>"},
			Wants:      `RevokeCertificates:\n{.*}`,
			DoesntWant: "Failed",
		},
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
	}
	resetTests := []Test{
		{
			Name:  "Resetting a Target Successfully",
			Args:  map[string]string{},
			Wants: `^$`,
		},
	}
	osTests := []Test{
		{
			Name:   "Compatible OS with Good Hash Install",
			Args:   map[string]string{"op": "install", "version": "&<version>", "os": "&<os_path>"},
			Wants:  `^$`,
			Prompt: []string{"version", "os_path"},
		},
		{
			Name:     "Install Already Installed OS",
			Args:     map[string]string{"op": "install", "version": "&<version>", "os": "&<os_path>"},
			MustFail: true,
			Wait:     3,
			Wants:    "OS version &<version> is already installed",
		},
		{
			Name:  "Activate Newly Installed OS",
			Args:  map[string]string{"op": "activate", "version": "&<version>"},
			Wants: `^$`,
		},
		{
			Name:  "Verify Newly Installed OS",
			Args:  map[string]string{"op": "verify", "version": "&<version>"},
			Wait:  20,
			Wants: "Running OS Version: &<version>",
		},
		{
			Name:     "Transfer an Incompatible OS",
			Args:     map[string]string{"op": "install", "version": "&<incompatible_version>", "os": "&<incompatible_os_path>"},
			MustFail: true,
			Wants:    "Failed Install: InstallError occurred: INCOMPATIBLE",
			Prompt:   []string{"incompatible_version", "incompatible_os_path"},
		},
		{
			Name:     "Transfer Already Running OS",
			Args:     map[string]string{"op": "install", "version": "&<version>", "os": "&<os_path>"},
			MustFail: true,
			Wants:    "Failed Install: InstallError occured: INSTALL_RUN_PACKAGE",
		},
		{
			Name:     "Parse OS Fails",
			Args:     map[string]string{"op": "install", "version": "&<bad_os_version>", "os": "&<bad_os_path>"},
			MustFail: true,
			Wants:    "Failed Install: InstallError occured: PARSE_FAIL",
			Prompt:   []string{"bad_os_version", "bad_os_path"},
		},
		{
			Name:     "OS Integrity Check Fails",
			Args:     map[string]string{"op": "install", "version": "&<bad_hash_version>", "os": "&<bad_hash_os>"},
			MustFail: true,
			Wants:    "Failed Install: InstallError occured: INTEGRITY_FAIL",
			Prompt:   []string{"bad_hash_version", "bad_hash_os"},
		},
		{
			Name:   "Activate Non Existent Version",
			Args:   map[string]string{"op": "activate", "version": "&<non_existent_version>"},
			Wants:  "Failed Activate: Non existent version: &<non_existent_version>",
			Prompt: []string{"non_existent_version"},
		},
	}
	return Tests{"gnoi_os": osTests, "gnoi_reset": resetTests, "gnoi_cert": certTests}
}
