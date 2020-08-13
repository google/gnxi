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
	"sync"
)

// Manager for storing data on OS's.
type Manager struct {
	osMap                 map[string]bool
	failMsgs              map[string]string
	runningVersion        string
	factoryVersion        string
	activationFailMessage string
	mu                    sync.RWMutex
}

// Settings wraps OS Server initialization options.
type Settings struct {
	FactoryVersion      string
	InstalledVersions   []string
	ReceiveChunkSizeAck uint64
}

// NewManager for OS service module. Will manage state of OS module.
func NewManager(factoryVersion string) *Manager {
	return &Manager{
		osMap:          map[string]bool{factoryVersion: true},
		factoryVersion: factoryVersion,
		failMsgs:       map[string]string{},
	}
}

// IsRunning will tell us whether or not the OS version specified is currently running.
func (m *Manager) IsRunning(version string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return version == m.runningVersion
}

// SetRunning sets the running OS to the version specified.
func (m *Manager) SetRunning(version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.activationFailMessage = m.failMsgs[version]; m.activationFailMessage != "" {
		return nil
	}
	if _, ok := m.osMap[version]; ok {
		m.runningVersion = version
		return nil
	}
	return fmt.Errorf("NON_EXISTENT_VERSION")
}

// Install installs an OS. It must be fully transferred and verified beforehand.
func (m *Manager) Install(version, activationFailMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.osMap[version] = true
	m.failMsgs[version] = activationFailMsg
}

// IsInstalled returns true if the OS is installed.
func (m *Manager) IsInstalled(version string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.osMap[version]
}
