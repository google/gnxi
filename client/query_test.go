/* Copyright 2017 Google Inc.

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
package client

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/proto/gnmi"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		query         string
		parsedQueries []string
	}{
		{
			query:         "/a/b/c/d",
			parsedQueries: []string{"/a/b/c/d"},
		},
		{
			query:         "a/b/c/d,c/d/e/f",
			parsedQueries: []string{"a/b/c/d", "c/d/e/f"},
		},
		{
			query:         "/a/b/c/d[12/3=4]/e",
			parsedQueries: []string{"/a/b/c/d[12/3=4]/e"},
		},
	}
	for _, tt := range tests {
		got := ParseQuery(tt.query)
		if diff := pretty.Compare(tt.parsedQueries, got); diff != "" {
			t.Errorf("ParseQuery(%s) returned diff (-want +got):\n%s", tt.query, diff)
		}
	}
}

func TestToGetRequest(t *testing.T) {
	tests := []struct {
		queries    []string
		getRequest *gnmi.GetRequest
	}{
		{
			queries: []string{"/", "/a/b/c/d[a=123]/e", "c/d[\"a/b\"=\"12/3\"]"},
			getRequest: &gnmi.GetRequest{
				Path: []*gnmi.Path{
					{
						Elem: []*gnmi.PathElem{},
					},
					{
						Elem: []*gnmi.PathElem{
							{Name: "a"},
							{Name: "b"},
							{Name: "c"},
							{Name: "d", Key: map[string]string{"a": "123"}},
							{Name: "e"},
						},
					},
					{
						Elem: []*gnmi.PathElem{
							{Name: "c"},
							{Name: "d", Key: map[string]string{"a/b": "12/3"}},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		got, err := ToGetRequest(tt.queries)
		if err != nil {
			t.Errorf("ToGetRequest(%s) returned error: %s", tt.queries, err)
		}
		if diff := pretty.Compare(tt.getRequest, got); diff != "" {
			t.Errorf("ToGetRequest(%s) returned diff (-want +got):\n%s", tt.queries, diff)
		}
	}
}
