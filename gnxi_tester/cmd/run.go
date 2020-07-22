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

package cmd

import (
	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/orchestrator"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run set of tests.",
	Long:    "Run a set of tests from the config file",
	Example: "gnxi_tester run [test_names]",
	Run:     handleRun,
}

// handleRun will run some or all of the tests.
func handleRun(cmd *cobra.Command, args []string) {
	if success, err := orchestrator.RunTests(args); err != nil {
		log.Exit(err)
	} else {
		log.Info("tests run successfully: %s", success)
	}
}
