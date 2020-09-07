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
	"path/filepath"

	"github.com/spf13/viper"
)

// Target stores connection details of a target.
type Target struct {
	Address string `json:"address" mapstructure:"address"`
	Ca      string `json:"ca" mapstructure:"ca"`
	CaKey   string `json:"cakey" mapstructure:"cakey"`
}

// SetTarget adds any new target to the list of known targets.
func SetTarget(targetName, targetAddress, ca, caKey string, abs bool) error {
	if err := prepareTarget(targetName, targetAddress, ca, caKey, abs); err != nil {
		return err
	}
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// prepareTarget parses provided details and creates or modifies target entry.
func prepareTarget(targetName, targetAddress, ca, caKey string, abs bool) error {
	var caPath string
	var caKeyPath string
	var err error
	targets := GetTargets()
	if targets == nil {
		targets = map[string]Target{}
	}
	if targetName == "" {
		if len(targets) > 0 {
			return nil
		}
		return errors.New("No targets in history and no target specified")
	}
	if abs {
		if ca != "" || caKey != "" {
			caPath, err = filepath.Abs(ca)
			if err != nil {
				return err
			}
			caKeyPath, err = filepath.Abs(caKey)
			if err != nil {
				return err
			}
		}
	} else {
		caPath, caKeyPath = ca, caKey
	}
	if _, exists := targets[targetName]; !exists {
		if targetAddress == "" || ca == "" || caKey == "" {
			return errors.New("Target not found")
		}
		targets[targetName] = Target{
			Address: targetAddress,
			Ca:      caPath,
			CaKey:   caKeyPath,
		}
	} else {
		target := targets[targetName]
		if targetAddress != "" {
			target.Address = targetAddress
		}
		if ca != "" {
			target.Ca = caPath
		}
		if caKey != "" {
			target.CaKey = caKeyPath
		}
		targets[targetName] = target
	}
	viper.Set("targets.last_target", targetName)
	viper.Set("targets.devices", targets)
	return nil
}
