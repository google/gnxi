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
	MustFail   bool              `mapstructure:"mustfail"`
	Wait       int               `mapstructure:"wait"`
	Wants      string            `mapstructure:"wants"`
	DoesntWant string            `mapstructure:"doesntwant"`
	Prompt     []string          `mapstructure:"prompt"`
}

func setDefaults() {
	testCases, order := generateTestCases()
	viper.SetDefault("tests", testCases)
	viper.SetDefault("docker.build", "golang:1.14-alpine")
	viper.SetDefault("docker.runtime", "alpine:latest")
	viper.SetDefault("docker.files", createDockerfiles(testCases))
	viper.SetDefault("order", order)
	viper.SetDefault("web.prompts", map[string]Prompts{})
	viper.SetDefault("files", map[string][]string{
		"gnoi_os": {"os_path", "new_os_path"},
	})
}

func generateTestCases() (Tests, []string) {
	provisionTest := []Test{{
		Name:   "Provision Bootstrapping Target",
		Args:   map[string]string{"op": "provision", "cert_id": "&<cert_id>"},
		Wants:  "Install success",
		Prompt: []string{"cert_id"},
	}}
	certTests := []Test{
		{
			Name:  "Get certs",
			Args:  map[string]string{"op": "get"},
			Wait:  10,
			Wants: `GetCertificates:\n{\w*`,
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
			Wants: `CanGenerateCSR:\n(true|false)`,
		},
		{
			Name:       "Revoke a cert",
			Args:       map[string]string{"op": "revoke", "cert_id": "&<new_cert_id>"},
			Wants:      `RevokeCertificates:\n{\w*`,
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
			Wants: `Reset Called Successfully!`,
		},
	}
	osTests := []Test{
		{
			Name:   "Compatible OS with Good Hash Install",
			Args:   map[string]string{"op": "install", "version": "&<os_version>", "os": "&<os_path>"},
			Wants:  `^$`,
			Prompt: []string{"os_version"},
		},
		{
			Name:  "Install Already Installed OS",
			Args:  map[string]string{"op": "install", "version": "&<os_version>", "os": "&<os_path>"},
			Wait:  3,
			Wants: "OS version &<os_version> is already installed",
		},
		{
			Name:  "Activate Newly Installed OS",
			Args:  map[string]string{"op": "activate", "version": "&<os_version>"},
			Wants: `^$`,
		},
		{
			Name:  "Verify Newly Installed OS",
			Args:  map[string]string{"op": "verify"},
			Wait:  20,
			Wants: "Running OS version: &<os_version>",
		},
		{
			Name:     "Force Transfer Already Running OS",
			Args:     map[string]string{"op": "install", "os": "&<os_path>"},
			MustFail: true,
			Wants:    "Failed Install: InstallError occurred: INSTALL_RUN_PACKAGE",
		},
		{
			Name:   "Install another OS",
			Args:   map[string]string{"op": "install", "version": "&<new_os_version>", "os": "&<new_os_path>"},
			Wants:  `^$`,
			Prompt: []string{"new_os_version"},
		},
		{
			Name:  "Activate Newly Installed OS",
			Args:  map[string]string{"op": "activate", "version": "&<new_os_version>"},
			Wants: `^$`,
		},
		{
			Name:  "Verify Newly Installed OS",
			Args:  map[string]string{"op": "verify"},
			Wait:  20,
			Wants: "Running OS version: &<new_os_version>",
		},
		{
			Name:     "Activate Non Existent Version",
			Args:     map[string]string{"op": "activate", "version": "&<non_existent_os_version>"},
			MustFail: true,
			Wants:    "Failed Activate: Non existent version: &<non_existent_os_version>",
			Prompt:   []string{"non_existent_os_version"},
		},
	}
	return Tests{"gnoi_os": osTests, "gnoi_reset": resetTests, "gnoi_cert": certTests, "provision": provisionTest}, []string{"gnoi_os", "gnoi_cert", "gnoi_reset"}
}
