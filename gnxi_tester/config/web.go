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

// Prompts represents a a prompt config set that will get stored in viper.
type Prompts struct {
	Name    string            `json:"name" mapstructure:"name"`
	Prompts map[string]string `json:"prompts" mapstructure:"name"`
	Files   map[string]string `json:"files" mapsstructure:"name"`
}

// Set prompts in viper.
func (p *Prompts) Set() error {
	prompts := viper.GetStringMap("web.prompts")
	prompts[p.Name] = *p
	viper.AllKeys()
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// GetPrompts returns a slice of all promts configs available.
func GetPrompts() []Prompts {
	out := []Prompts{}
	webPrompts := viper.GetStringMap("web.prompts")
	for _, p := range webPrompts {
		if prompts, ok := p.(Prompts); ok {
			out = append(out, prompts)
		}
	}
	return out
}
