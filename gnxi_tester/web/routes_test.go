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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
)

type routerPath struct {
	Path    string
	Methods []string
}

func TestGenerateRouter(t *testing.T) {
	r := generateRouter()
	want := []routerPath{
		{"/prompts", []string{"GET"}},
		{"/prompts/list", []string{"GET"}},
		{"/prompts", []string{"POST", "PUT", "OPTIONS"}},
		{"/prompts/{name}", []string{"DELETE", "OPTIONS"}},
		{"/target", []string{"GET"}},
		{"/target/{name}", []string{"GET"}},
		{"/target/{name}", []string{"POST", "PUT", "OPTIONS"}},
		{"/target/{name}", []string{"DELETE"}},
		{"/file", []string{"POST"}},
		{"/file/{file}", []string{"DELETE", "OPTIONS"}},
		{"/run", []string{"POST", "OPTIONS"}},
		{"/run/output", []string{"GET"}},
		{"/test", []string{"GET"}},
		{"/test/order", []string{"GET"}},
	}
	got := []routerPath{}
	r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			t.Errorf("Unexpected error walking routes: %v", err)
		}
		methods, err := route.GetMethods()
		if err != nil {
			t.Errorf("Unexpected error getting methods: %v", err)
		}
		got = append(got, routerPath{
			path,
			methods,
		})
		return nil
	})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected routes found (-want +got): %s.", diff)
	}
}
