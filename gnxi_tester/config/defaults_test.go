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
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestGetDefaults(t *testing.T) {
	got := generateTestCases()
	if got == nil {
		t.Errorf("Reset Tests Not Set!")
	}
	if got["gnoi_os"] == nil {
		t.Errorf("OS Tests Not Set!")
	}
	if got["gnoi_cert"] == nil {
		t.Errorf("Certificate Tests Not Set!")
	}
	if got["gnoi_reset"] == nil {
		t.Errorf("Reset Tests Not Set!")
	}
}

func TestSetDefaults(t *testing.T) {
	want := generateTestCases()
	setDefaults()
	got := GetTests()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("GetTests(): (-got +want):\n%s", diff)
	}
}
