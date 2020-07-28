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
	"fmt"
	"regexp"
	"strings"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/spf13/viper"
)

type callbackFunc func(name string) string

const (
	openDelim  = "&<"
	closeDelim = ">"
)

var (
	configTests map[string][]config.Test
	input       = map[string]string{}
	delimRe     = regexp.MustCompile(fmt.Sprintf("%s.*%s", openDelim, closeDelim))
)

// RunTests will take in test name and run each test or all tests.
func RunTests(tests []string, prompt callbackFunc) (success []string, err error) {
	var output string
	if len(tests) == 0 {
		configTests = config.GetTests()
		for name := range configTests {
			if output, err = runTest(name, prompt); err != nil {
				return
			}
			success = append(success, output)
		}
	} else {
		for _, name := range tests {
			if output, err = runTest(name, prompt); err != nil {
				return
			}
			success = append(success, output)
		}
	}
	return
}

func runTest(name string, prompt callbackFunc) (string, error) {
	tests := configTests[name]
	targetName := viper.GetString("targets.last_target")
	defaultArgs := fmt.Sprintf(
		"-logtostderr -target_name %s -target_addr %s",
		targetName,
		viper.GetStringMap("targets.devices")[targetName],
	)
	stdout := fmt.Sprintf("*%s*:", name)
	for _, test := range tests {
		for _, p := range test.Prompt {
			input[p] = prompt(p)
		}
		test.DoesntWant = insertVars(test.DoesntWant)
		test.Wants = insertVars(test.Wants)
		binArgs := defaultArgs
		for arg, val := range test.Args {
			binArgs = fmt.Sprintf("-%s %s %s", arg, insertVars(val), binArgs)
		}
		out, code, err := RunContainer(name, binArgs)
		if exp := expects(out, &test); (code == 0) == test.MustFail || err != nil || exp != nil {
			return "", formateErr(name, test.Name, exp, code, test.MustFail, out, err)
		}
		stdout = fmt.Sprintf("%s\n%s:\n%s\n", stdout, test.Name, out)
	}
	return stdout, nil
}

func expects(out string, test *config.Test) error {
	if len(test.Wants) > 0 {
		wantsRe := regexp.MustCompile(test.Wants)
		if i := wantsRe.FindStringIndex(out); i == nil {
			return fmt.Errorf("Wanted %s in output", test.Wants)
		}
	}
	if len(test.DoesntWant) > 0 {
		doesntRe := regexp.MustCompile(test.DoesntWant)
		if i := doesntRe.FindStringIndex(out); i != nil {
			return fmt.Errorf("Didn't want %s in output", test.DoesntWant)
		}
	}
	return nil
}

func formateErr(major, minor string, custom error, code int, fail bool, out string, err error) error {
	return fmt.Errorf(
		"Error occured in test %s-%s: (%v) exitCode(%d), mustFail(%v), stdout(%s), runtimeErr(%v)",
		major,
		minor,
		custom,
		code,
		fail,
		out,
		err,
	)
}

func insertVars(in string) string {
	matches := delimRe.FindAllString(in, -1)
	for _, match := range matches {
		in = strings.Replace(in, match, input[match[2:len(match)-1]], 1)
	}
	return in
}
