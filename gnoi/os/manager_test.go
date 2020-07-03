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
	t.Run("Is running after set running", func(t *testing.T) {
		manager := NewManager("new")
		manager.runningVersion = "new"
		if !manager.IsRunning("new") {
			t.Error("OS not running after being set running")
		}
	})
	t.Run("Is not running", func(t *testing.T) {
		manager := NewManager("new")
		if manager.IsRunning("new") {
			t.Error("OS running but not set running")
		}
	})
}

func TestSetRunning(t *testing.T) {
	t.Run("Is running if OS exists", func(t *testing.T) {
		manager := NewManager("new")
		err := manager.SetRunning("new")
		if err != nil {
			t.Errorf("Error running OS: %w", err)
		}
	})
	t.Run("Is not running if OS doesn't exists", func(t *testing.T) {
		manager := NewManager("otherNew")
		err := manager.SetRunning("new")
		if err == nil {
			t.Errorf("OS is running despite not existing")
		}
	})
}

func TestInstall(t *testing.T) {
	t.Run("Is installed", func(t *testing.T) {
		manager := NewManager("new")
		manager.Install("newer")
		if _, ok := manager.osMap["newer"]; !ok {
			t.Error("OS not installed")
		}
	})
}
