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

package os

import (
	"fmt"
)

// Manager for storing data on OS's.
type Manager struct {
	osMap          map[string]bool
	failMsgs       map[string]string
	runningVersion string
	factoryVersion string
}

// Settings wraps OS Server initialization options.
type Settings struct {
	FactoryVersion    string
	InstalledVersions []string
}

// NewManager for OS service module. Will manage state of OS module.
func NewManager(factoryVersion string) *Manager {
	return &Manager{
		osMap:          map[string]bool{factoryVersion: true},
		factoryVersion: factoryVersion,
	}
}

// IsRunning will tell us whether or not the OS version specified is currently running.
func (m *Manager) IsRunning(version string) bool {
	return version == m.runningVersion
}

// SetRunning sets the running OS to the version specified.
func (m *Manager) SetRunning(version string) error {
	if _, ok := m.osMap[version]; ok {
		m.runningVersion = version
		return nil
	}
	return fmt.Errorf("NON_EXISTENT_VERSION")
}

// Install installs an OS. It must be fully transfered and verified beforehand.
func (m *Manager) Install(version string) {
	m.osMap[version] = true
}
