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
	"path"
	"regexp"
	"strings"

	log "github.com/golang/glog"
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
	files       map[string]string
)

// RunTests will take in test name and run each test or all tests.
func RunTests(tests []string, prompt callbackFunc, userFiles map[string]string) (success []string, err error) {
	files = userFiles
	defaultOrder := viper.GetStringSlice("order")
	if err = InitContainers(defaultOrder); err != nil {
		return
	}
	configTests = config.GetTests()
	var output string
	if len(tests) == 0 {
		defaultOrder = viper.GetStringSlice("order")
		provisionTests := configTests["provision"]
		if output, err = runTest("gnoi_cert", prompt, provisionTests); err != nil {
			return
		}
		for _, name := range defaultOrder {
			if output, err = runTest(name, prompt, configTests[name]); err != nil {
				return
			}
			success = append(success, output)
		}
	} else {
		for _, name := range tests {
			if output, err = runTest(name, prompt, configTests[name]); err != nil {
				return
			}
			success = append(success, output)
		}
	}
	return
}

func runTest(name string, prompt callbackFunc, tests []config.Test) (string, error) {
	log.Infof("Running major test %s", name)
	targetName := viper.GetString("targets.last_target")
	target := config.GetDevices()[targetName]
	defaultArgs := fmt.Sprintf(
		"-logtostderr -target_name %s -target_addr %s -ca /certs/ca.crt -ca_key /certs/ca.key",
		targetName,
		target.Address,
	)
	stdout := fmt.Sprintf("*%s*:", name)
	for _, test := range tests {
		log.Infof("Running minor test %s:%s", name, test.Name)
		for _, p := range test.Prompt {
			input[p] = prompt(p)
		}
		test.DoesntWant = insertVars(test.DoesntWant, &[]string{})
		test.Wants = insertVars(test.Wants, &[]string{})
		binArgs := defaultArgs
		insertFiles := []string{}
		for arg, val := range test.Args {
			binArgs = fmt.Sprintf("-%s %s %s", arg, insertVars(val, &insertFiles), binArgs)
		}
		out, code, err := RunContainer(name, binArgs, &target, insertFiles)
		if exp := expects(out, &test); (code == 0) == test.MustFail || err != nil || exp != nil {
			return "", formatErr(name, test.Name, out, exp, code, test.MustFail, binArgs, err)
		}
		stdout = fmt.Sprintf("%s\n%s:\n%s\n", stdout, test.Name, out)
		log.Infof("Successfully run test %s:%s", name, test.Name)
	}
	return stdout, nil
}

func expects(out string, test *config.Test) error {
	if len(test.Wants) > 0 {
		wantsRe := regexp.MustCompile(test.Wants)
		if i := wantsRe.FindStringIndex(out); i == nil {
			return fmt.Errorf("Wanted '%s' in output", test.Wants)
		}
	}
	if len(test.DoesntWant) > 0 {
		doesntRe := regexp.MustCompile(test.DoesntWant)
		if i := doesntRe.FindStringIndex(out); i != nil {
			return fmt.Errorf("Didn't want '%s' in output", test.DoesntWant)
		}
	}
	return nil
}

func formatErr(major, minor, out string, custom error, code int, fail bool, args string, err error) error {
	return fmt.Errorf(
		"Error occured in test %s-<%s>: \nwantedErr(%v)\nexitCode(%d)\nmustFail(%v)\ndaemonErr(%v)\nargs(%s)\noutput:\n# %s)",
		major,
		minor,
		custom,
		code,
		fail,
		err,
		args,
		out,
	)
}

func insertVars(in string, insertFiles *[]string) string {
	matches := delimRe.FindAllString(in, -1)
	for _, match := range matches {
		name := match[2 : len(match)-1]
		if f, ok := files[name]; ok {
			in = strings.Replace(in, match, path.Join("/tmp", path.Base(f)), 1)
			*insertFiles = append(*insertFiles, f)
		} else {
			in = strings.Replace(in, match, input[name], 1)
		}
	}
	return in
}
