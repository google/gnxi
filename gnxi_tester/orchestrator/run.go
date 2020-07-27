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

package orchestrator

import (
	"github.com/google/gnxi/gnxi_tester/config"
)

// RunTests will take in test name and run each test or all tests.
func RunTests(tests []string) (success []string, err error) {
	var output string
	if len(tests) == 0 {
		configTests := config.GetTests()
		for name := range configTests {
			if output, err = runTest(name); err != nil {
				return
			}
			success = append(success, output)
		}
	} else {
		for _, name := range tests {
			if output, err = runTest(name); err != nil {
				return
			}
			success = append(success, output)
		}
	}
	return
}

func runTest(test string) (string, error) {
	return "", nil
}
