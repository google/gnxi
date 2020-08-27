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

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestSetPrompts(t *testing.T) {
	viper.SetConfigFile("/tmp/config.yml")
	t.Run("will set", func(t *testing.T) {
		prompts := &Prompts{"test", map[string]string{}, map[string]string{}}
		if err := prompts.Set(); err != nil {
			t.Error(err)
		}
		var got Prompts
		if err := viper.UnmarshalKey("web.prompts.test", &got); err != nil {
			t.Errorf("Couldn't find prompts in map.")
		}
		if diff := cmp.Diff(*prompts, got); diff != "" {
			t.Errorf("Didn't get correct prompts (-want +got) %s.", diff)
		}
	})
}
func TestGetPrompts(t *testing.T) {
	viper.SetConfigFile("/tmp/config.yml")
	t.Run("will get", func(t *testing.T) {
		viper.Set("web.prompts", map[string]interface{}{})
		prompts := &Prompts{"test", map[string]string{}, map[string]string{}}
		if err := prompts.Set(); err != nil {
			t.Errorf("Error occured in setting prompts: %w", err)
		}
		got := GetPrompts()
		if len(got) == 0 {
			t.Fatal("No prompts received.")
		}
		if diff := cmp.Diff(*prompts, got["test"]); diff != "" {
			t.Errorf("Didn't get correct prompts (-want +got) %s.", diff)
		}
	})
}
