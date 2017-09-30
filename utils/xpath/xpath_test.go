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

package xpath

import (
	"reflect"
	"testing"
)

func TestSplitPath(t *testing.T) {
	tests := []struct {
		desc     string
		path     string
		expectOK bool
		want     []string
	}{{
		desc:     "test empty path string",
		path:     "",
		expectOK: true,
	}, {
		desc:     "test path without List path element",
		path:     "/a/b/c",
		expectOK: true,
		want:     []string{"a", "b", "c"},
	}, {
		desc:     "test path containing extra trailing /",
		path:     "/a/b/c/",
		expectOK: true,
		want:     []string{"a", "b", "c"},
	}, {
		desc:     "test path containing extra leading /",
		path:     "//a/b/c",
		expectOK: true,
		want:     []string{"a", "b", "c"},
	}, {
		desc:     "test path containing extra / in the middle",
		path:     "/a//b/c",
		expectOK: true,
		want:     []string{"a", "b", "c"},
	}, {
		desc:     "test path without leading /",
		path:     "a/b/c",
		expectOK: true,
		want:     []string{"a", "b", "c"},
	}, {
		desc:     "test path containing List element without / inside",
		path:     "/a/b[k1=10]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=10]", "c"},
	}, {
		desc:     "test path containing List element without / inside, with [ inside key value pair string",
		path:     "/a/b[k1=[10]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=[10]", "c"},
	}, {
		desc:     `test path containing List element without / inside, with \] inside key value pair string`,
		path:     `/a/b[k1=\]10]/c`,
		expectOK: true,
		want:     []string{"a", `b[k1=\]10]`, "c"},
	}, {
		desc:     "test path containing List element without / inside, with ] after key value pair string",
		path:     "/a/b[k1=]10]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=]10]", "c"},
	}, {
		desc:     "test path containing List element without / inside, with ] before key value pair string",
		path:     "/a/b][k1=10]/c",
		expectOK: true,
		want:     []string{"a", "b][k1=10]", "c"},
	}, {
		desc:     "test path containing List element without / inside, with ] in non List path element",
		path:     "/a]/b[k1=10]/c",
		expectOK: true,
		want:     []string{"a]", "b[k1=10]", "c"},
	}, {
		desc:     "test path containing List element without / inside, missing [",
		path:     "/a/bk1=10]/c",
		expectOK: true,
		want:     []string{"a", "bk1=10]", "c"},
	}, {
		desc: "test path containing List element without / inside, missing ]",
		path: "/a/b[k1=10/c",
	}, {
		desc:     "test path containing List element with / inside",
		path:     "/a/b[k1=10.10.10.10/24]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=10.10.10.10/24]", "c"},
	}, {
		desc:     "test path containing List element with / inside, extra [ before /",
		path:     "/a/b[k1=10.10.[10.10/24]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=10.10.[10.10/24]", "c"},
	}, {
		desc:     "test path containing List element with / inside, extra ] after /",
		path:     "/a/b[k1=10.10.10.10/24]]/c",
		expectOK: true,
		want:     []string{"a", "b[k1=10.10.10.10/24]]", "c"},
	}, {
		desc:     "test path containing List element with / inside, missing [",
		path:     "/a/bk1=10.10.10.10/24]/c",
		expectOK: true,
		want:     []string{"a", "bk1=10.10.10.10", "24]", "c"},
	}, {
		desc: "test path containing List element with / inside, missing ]",
		path: "/a/b[k1=10.10.10.10/24/c",
	}}

	for _, test := range tests {
		got, err := splitPath(test.path)
		if test.expectOK {
			if err != nil {
				t.Errorf("%s: splitPath(%q) got error: %v, wanted error: %v",
					test.desc, test.path, err, !test.expectOK)
				continue
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%s: splitPath(%q) got: %v, wanted: %v",
					test.desc, test.path, got, test.want)
			}
		} else if err == nil {
			t.Errorf("%s: splitPath(%q) got error: %v, wanted error: %v",
				test.desc, test.path, err, !test.expectOK)
		}
	}
}

func TestParseElement(t *testing.T) {
	tests := []struct {
		desc     string
		elem     string
		expectOK bool
		want     []interface{}
	}{{
		desc:     "test non-List node name success",
		elem:     "a-b_c0",
		expectOK: true,
		want:     []interface{}{"a-b_c0"},
	}, {
		desc: "test empty string",
		elem: "",
	}, {
		desc: "test invalid non-List node name",
		elem: "a-b]c0",
	}, {
		desc: "test single-key List key value pair, with [ in key leaf name",
		elem: "a[bc[k1=v1]",
	}, {
		desc: "test single-key List key value pair, with ] in key leaf name",
		elem: "a[bc]k1=v1]",
	}, {
		desc: "test single-key List key value pair, with extra letters after ]",
		elem: "a[k1=v1]abc",
	}, {
		desc: "test key value pair without =",
		elem: "a[k1v1]",
	}, {
		desc: "test key value pair without key leaf name",
		elem: "a[=abcde]",
	}, {
		desc: "test key value pair without key leaf value",
		elem: "a[key-name=]",
	}, {
		desc: "test key value pair string containing only =",
		elem: "a[=]",
	}, {
		desc: "test key value pair string containing nonthing",
		elem: "a[]",
	}, {
		desc: "test single-key List key value pair, ] missing",
		elem: "a[k1=v1",
	}, {
		desc: "test single-key List key value pair, [ missing",
		elem: "ak1=v1]",
	}, {
		desc:     "test single-key List key value pair success",
		elem:     "a[k1=v1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v1"}},
	}, {
		desc:     `test single-key List key value pair success, with \] in key value`,
		elem:     `a[k1=v\]1]`,
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v]1"}},
	}, {
		desc:     `test single-key List key value pair success, with \[ in key value`,
		elem:     `a[k1=v\[1]`,
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v[1"}},
	}, {
		desc: `test single-key List key value pair success, with ] in key value`,
		elem: `a[k1=v]1]`,
	}, {
		desc:     `test single-key List key value pair success, with [ in key value`,
		elem:     `a[k1=v[1]`,
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v[1"}},
	}, {
		desc:     `test single-key List key value pair success, with \\ in key value`,
		elem:     "a[k1=v\\1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": `v\1`}},
	}, {
		desc:     `test single-key List key value pair success, with \" in key value`,
		elem:     "a[k1=v\"1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": `v"1`}},
	}, {
		desc:     `test single-key List key value pair success, with ' in key value`,
		elem:     "a[k1=v'1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": `v'1`}},
	}, {
		desc:     "test single-key List key value pair success, with ` in key value",
		elem:     "a[k1=v`1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v`1"}},
	}, {
		desc:     "test single-key List key value pair success, with = in key value",
		elem:     "a[k1=v=1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v=1"}},
	}, {
		desc:     "test single-key List key value pair success, with \n in key value",
		elem:     "a[k1=v\n1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v\n1"}},
	}, {
		desc:     "test single-key List key value pair success, with \r in key value",
		elem:     "a[k1=v\r1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v\r1"}},
	}, {
		desc:     "test single-key List key value pair success, with \t in key value",
		elem:     "a[k1=v\t1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v\t1"}},
	}, {
		desc:     "test single-key List key value pair success, with \x12 in key value",
		elem:     "a[k1=v\x121]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v\x121"}},
	}, {
		desc:     "test single-key List key value pair success, with unicode character \u212A in key value",
		elem:     "a[k1=v\u212A1]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v\u212A1"}},
	}, {
		desc: "test multi-key List key value pair string, missing [ in second key-value pair",
		elem: "a[k1=v1]k2=v2]",
	}, {
		desc: "test multi-key List key value pair string, missing ] in second key-value pair",
		elem: "a[k1=v1][k2=v2",
	}, {
		desc: "test multi-key List key value pair string, with unescaped [ in second key value",
		elem: "a[k1=v1][k2=v2]]",
	}, {
		desc: "test multi-key List key value pair string, with unescaped [ in second key leaf name",
		elem: "a[k1=v1][[k2=v2]",
	}, {
		desc: "test multi-key List key value pair string with extra letters between key-value pairs",
		elem: "a[k1=v1]asdf[k2=v2]",
	}, {
		desc:     "test multi-key List key value pair string",
		elem:     "a[k1=v1][k2=v2]",
		expectOK: true,
		want:     []interface{}{"a", map[string]string{"k1": "v1", "k2": "v2"}},
	}}

	for _, test := range tests {
		got, err := parseElement(test.elem)
		if test.expectOK {
			if err != nil {
				t.Errorf("%s: parseElement(%q) got error: %v, wanted error: %v",
					test.desc, test.elem, err, !test.expectOK)
				continue
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%s: parseElement(%q) got: %v, wanted: %v",
					test.desc, test.elem, got, test.want)
			}
		} else if err == nil {
			t.Errorf("%s: parseElement(%q) got error: %v, wanted error: %v",
				test.desc, test.elem, err, !test.expectOK)
		}
	}
}

func TestParseStringPath(t *testing.T) {
	tests := []struct {
		desc     string
		path     string
		expectOK bool
		want     []interface{}
	}{{
		desc:     "test empty path string",
		path:     "",
		expectOK: true,
	}, {
		desc:     "test path without List element",
		path:     "/a/b/c",
		expectOK: true,
		want:     []interface{}{"a", "b", "c"},
	}, {
		desc:     "test path containing a single-key List",
		path:     "/a/b[k1=10]/c",
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "10"}, "c"},
	}, {
		desc: "test path containing a single-key List, invalid List name",
		path: `/a/b\[k1=10]/c`,
	}, {
		desc:     "test path containing a single-key List with / in key leaf value",
		path:     "/a/b[k1=10.10.10.10/24]/c",
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "10.10.10.10/24"}, "c"},
	}, {
		desc:     "test path containing a single-key List with [ in key leaf value",
		path:     `/a/b[k1=10.10.10.10\[24]/c`,
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "10.10.10.10[24"}, "c"},
	}, {
		desc:     "test path containing a single-key List with ] in key leaf value",
		path:     "/a/b[k1=10.10.10.10\\]24]/c",
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "10.10.10.10]24"}, "c"},
	}, {
		desc:     "test path containing multiple Lists",
		path:     "/a/b[k1=v1]/c/d[k2=v2]/e",
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "v1"}, "c", "d", map[string]string{"k2": "v2"}, "e"},
	}, {
		desc:     "test path containing a multi-key List",
		path:     `/a/b[k1=exact][k2=10.10.10.10/24]/c`,
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "exact", "k2": "10.10.10.10/24"}, "c"},
	}, {
		desc:     `test path containing a multi-key List with \][ in key leaf value`,
		path:     `/a/b[k1=10\][][k2=abc]/c`,
		expectOK: true,
		want:     []interface{}{"a", "b", map[string]string{"k1": "10][", "k2": "abc"}, "c"},
	}, {
		desc: "test path containing a multi-key List but missing ] in second key-value string",
		path: "/a/b[k1=10][k2=abc/c",
	}, {
		desc: "test path containing a multi-key List with unescaped [ in second key leaf name",
		path: "/a/b[k1=10][[k2=abc]/c",
	}, {
		desc: "test path containing a multi-key List, second key-value pair without [ and ]",
		path: "/a/b[k1=10]k2=abc/c",
	}}

	for _, test := range tests {
		got, err := ParseStringPath(test.path)
		if test.expectOK {
			if err != nil {
				t.Errorf("%s: ParseStringPath(%q) got error: %v, wanted error: %v",
					test.desc, test.path, err, !test.expectOK)
				continue
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%s: ParseStringPath(%q) got: %v, wanted: %v",
					test.desc, test.path, got, test.want)
			}
		} else if err == nil {
			t.Errorf("%s: ParseStringPath(%q) got error: %v, wanted error: %v",
				test.desc, test.path, err, !test.expectOK)
		}
	}
}
