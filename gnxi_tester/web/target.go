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

package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func getNameParam(w http.ResponseWriter, r *http.Request) string {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		logErr(r.Header, errors.New("name param not set"))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return ""
	}
	return name
}

func handleTargetGet(w http.ResponseWriter, r *http.Request) {
	name := getNameParam(w, r)
	if name == "" {
		return
	}
	devices := config.GetDevices()
	device, ok := devices[name]
	if !ok {
		logErr(r.Header, errors.New("device not found"))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(device); err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleTargetSet(w http.ResponseWriter, r *http.Request) {
	name := getNameParam(w, r)
	if name == "" {
		return
	}
	devices := config.GetDevices()
	device := config.Device{}
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	devices[name] = device
	viper.Set("targets.devices", devices)
}
