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

import "testing"

func TestIsRunning(t *testing.T) {
	tests := []struct {
		name string
		running,
		wantsRun bool
	}{
		{"Is running after set running", true, false},
		{"Is not running", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := NewManager("new")
			if test.running {
				manager.runningVersion = "new"
			}
			if manager.IsRunning("new") == test.wantsRun {
				t.Error("OS not running after being set running")
			}
		})
	}
}

func TestSetRunning(t *testing.T) {
	tests := []struct {
		name,
		version string
		wantsErr bool
	}{
		{"Is running if OS exists", "new", false},
		{"Is not running if OS doesn't exists", "otherNew", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := NewManager("new")
			err := manager.SetRunning(test.version)
			if (err == nil) == test.wantsErr {
				t.Errorf("Error running OS: Error is %w", err)
			}
		})
	}
}

func TestInstall(t *testing.T) {
	t.Run("Is installed", func(t *testing.T) {
		manager := NewManager("new")
		manager.Install("newer", "")
		if _, ok := manager.osMap["newer"]; !ok {
			t.Error("OS not installed")
		}
	})
}
