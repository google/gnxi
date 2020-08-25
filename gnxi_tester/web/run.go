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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/gnxi/gnxi_tester/orchestrator"
	"github.com/spf13/viper"
)

const endTimeout = 5 * time.Second

var (
	outputBuffer = bytes.NewBuffer([]byte{})
	mu           = sync.Mutex{}
)

type runRequest struct {
	Prompts string   `json:"prompts"`
	Device  string   `json:"device"`
	Tests   []string `json:"tests"`
}

func writeToBuffer(target string, args ...interface{}) {
	outputBuffer.WriteString(fmt.Sprintf(target+"\n", args...))
}

var runTests = func(prompts config.Prompts, request runRequest) {
	promptHandler := func(name string) string {
		return prompts.Prompts[name]
	}
	mu.Lock()
	files := map[string]string{}
	dir := filesDir()
	for name, file := range prompts.Files {
		files[name] = path.Join(dir, file)
	}
	success, err := orchestrator.RunTests(request.Tests, promptHandler, files, writeToBuffer)
	if err != nil {
		outputBuffer.WriteString("\n" + err.Error())
	} else {
		for _, output := range success {
			outputBuffer.WriteString(output + "\n")
		}
	}
	<-time.After(endTimeout)
	outputBuffer.Reset()
	outputBuffer.WriteString("E0F")
	mu.Unlock()
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	request := runRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	configPrompts := viper.GetStringMap("web.prompts")
	promptsGeneric, ok := configPrompts[request.Prompts]
	if !ok {
		logErr(r.Header, fmt.Errorf("%s prompts not found", request.Prompts))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	prompts, ok := promptsGeneric.(config.Prompts)
	if !ok {
		logErr(r.Header, fmt.Errorf("%s prompts invalid", request.Prompts))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if _, ok := config.GetDevices()[request.Device]; !ok {
		logErr(r.Header, fmt.Errorf("%s device not found", request.Device))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	viper.Set("targets.last_target", request.Device)
	go runTests(prompts, request)
}

func handleRunOutput(w http.ResponseWriter, r *http.Request) {
	out, err := ioutil.ReadAll(outputBuffer)
	if err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(out)
}
