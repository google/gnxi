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
	"path"

	log "github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Init will read and if needed, initialize the config file.
func Init(filePath string) {
	if filePath == "" {
		home, err := homedir.Dir()
		if err != nil {
			log.Exitf("couldn't get home directory: %v", err)
		}
		filePath = path.Join(home, ".gnxi.yml")
	}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filePath)
	setDefaults()
	if err := viper.SafeWriteConfigAs(filePath); err != nil && err != viper.ConfigFileAlreadyExistsError(filePath) {
		log.Exitf("couldn't write config: %v", err)
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Exitf("couldn't read from config: %v", err)
	}
}

// GetTests will return tests from viper store.
func GetTests() Tests {
	var tests Tests
	if err := viper.UnmarshalKey("tests", &tests); err != nil {
		return nil
	}
	return tests
}

// GetDevices will return target connection history from Viper store.
func GetDevices() map[string]Device {
	var devices map[string]Device
	viper.UnmarshalKey("targets.devices", &devices)
	return devices
}
