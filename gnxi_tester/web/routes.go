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
	"net/http"

	log "github.com/golang/glog"
	"github.com/gorilla/mux"
)

// InitRouter will setup required routes for the web API and listen on the address.
func InitRouter(laddr string) {
	r := mux.NewRouter()

	r.HandleFunc("/prompts", handlePromptsGet).Methods("GET")
	r.HandleFunc("/prompts/list", handlePromptsList).Methods("GET")
	r.HandleFunc("/prompts", handlePromptsSet).Methods("POST", "PUT")
	r.HandleFunc("/target/{name}", handleTargetGet).Methods("GET")
	r.HandleFunc("/target/{name}", handleTargetSet).Methods("POST", "PUT")
	r.HandleFunc("/file", handleFileUpload).Methods("POST")
	r.HandleFunc("/file/{file}", handleFileDelete).Methods("DELETE")
	r.HandleFunc("/run", handleRun).Methods("POST")
	r.HandleFunc("/run/output", handleRunOutput).Methods("POST")

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req)
		})
	})

	log.Infof("Running web API on %s", laddr)

	if err := http.ListenAndServe(laddr, r); err != nil {
		log.Exit(err)
	}
}

var logErr = func(head http.Header, err error) {
	log.Errorf("error occured in view %v: %v", head, err)
}
