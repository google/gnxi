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
	"bytes"
	"fmt"
	"time"
)

// OS represents the data for a singular operating system.
type OS struct {
	Name,
	Hash,
	Version string
	BuildDate time.Time
	Data      *bytes.Buffer
}

// Manager for storing data on OS's.
type Manager struct {
	osMap             map[string]*OS
	running           *OS
	factory           *OS
	supportedVersions map[string]bool
	standbyState      StandbyState
}

// StandbyState of system.
type StandbyState string

const (
	unspecified StandbyState = "UNSPECIFIED"
	unsupported              = "UNSUPPORTED"
	nonExistent              = "NON_EXISTENT"
	unavailable              = "UNAVAILABLE"
)

// NewManager for OS service module. Will manage state of OS module.
func NewManager(supported map[string]bool, factoryOs *OS) *Manager {
	return &Manager{
		standbyState:      unsupported,
		supportedVersions: supported,
		factory:           factoryOs,
		osMap:             map[string]*OS{factoryOs.Version: factoryOs},
	}
}

// IsRunning will tell us whether or not the OS version specified is currently running.
func (m *Manager) IsRunning(version string) bool {
	if m.running != nil {
		return m.running.Version == version
	}
	return false
}

// SetRunning sets the running OS to the version specified.
func (m *Manager) SetRunning(version string) error {
	if o, ok := m.osMap[version]; ok {
		m.running = o
	}
	return fmt.Errorf("NON_EXISTENT_VERSION")
}

// Install an OS. Must be fully transfered and verified beforehand.
func (m *Manager) Install(o *OS) error {
	if supported := m.supportedVersions[o.Version]; !supported {
		return fmt.Errorf("INCOMPATIBLE")
	}
	if _, ok := m.osMap[o.Version]; !ok {
		m.osMap[o.Version] = o
	}
	return nil
}

// Rollback to factory OS and wipes all other OS's.
func (m *Manager) Rollback() error {
	for _, o := range m.osMap {
		if o.Version == m.factory.Version {
			continue
		}
		delete(m.osMap, o.Version)
	}
	m.running = m.factory
	return nil
}
