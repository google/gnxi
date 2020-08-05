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
	"bufio"
	"fmt"
	"os"

	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/gnxi/gnxi_tester/orchestrator"
	"github.com/spf13/cobra"
)

var (
	targetName    string
	targetAddress string
	ca            string
	caKey         string
	runCmd        = &cobra.Command{
		Use:     "run",
		Short:   "Run set of tests.",
		Long:    "Run a set of tests from the config file",
		Example: "gnxi_tester run [test_names]",
		Run:     handleRun,
	}
	scanner = bufio.NewReader(os.Stdin)
)

func init() {
	runCmd.Flags().StringVarP(&targetName, "target_name", "n", "", "The name of the target to be tested")
	runCmd.Flags().StringVarP(&targetAddress, "target_address", "a", "", "The address of the target to be tested")
	runCmd.Flags().StringVarP(&ca, "ca", "c", "", "The ca ")
	runCmd.Flags().StringVarP(&caKey, "ca_key", "k", "", "The name of the target to be tested")
}

// handleRun will run some or all of the tests.
func handleRun(cmd *cobra.Command, args []string) {
	if err := config.SetTarget(targetName, targetAddress, ca, caKey); err != nil {
		log.Exitf("Error writing config: %v", err)
	}
	if success, err := orchestrator.RunTests(args, promptUser); err != nil {
		log.Exitf("Error running tests: %v", err)
	} else {
		log.Infof("Tests ran successfully: %s", success)
	}
}

func promptUser(name string) string {
	fmt.Printf("Please provide %s: ", name)
	out, err := scanner.ReadString('\n')
	if err != nil {
		log.Exitf("error reading line: %v", err)
	}
	return out
}
