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
	"path/filepath"
	"strings"

	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/gnxi/gnxi_tester/orchestrator"
	"github.com/spf13/cobra"
)

var (
	targetName,
	targetAddress,
	ca,
	caKey,
	files string
	runCmd = &cobra.Command{
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
	runCmd.Flags().StringVarP(&caKey, "ca_key", "k", "", "The key for the ca file")
	runCmd.Flags().StringVarP(&files, "files", "f", "", "Extra files used for tests. Example: -f \"os_path:/path/to/os file:/path/to/other/file\"")
}

// handleRun will run some or all of the tests.
func handleRun(cmd *cobra.Command, args []string) {
	if err := config.SetTarget(targetName, targetAddress, ca, caKey); err != nil {
		log.Exitf("Error writing config: %v", err)
	}
	if success, err := orchestrator.RunTests(args, promptUser, parseFiles(), log.Infof); err != nil {
		log.Exitf("Error running tests: %v", err)
	} else {
		log.Info("Tests ran successfully:")
		for _, output := range success {
			log.Info(output)
		}
	}
}

func parseFiles() map[string]string {
	out := map[string]string{}
	for _, file := range strings.Split(files, " ") {
		file = strings.ReplaceAll(file, "\"", "")
		if len(file) > 1 {
			vals := strings.SplitN(file, ":", 2)
			if len(vals) > 1 {
				p, err := filepath.Abs(vals[1])
				if err != nil {
					log.Errorf("Filepath %s invalid", file)
				}
				out[vals[0]] = p
			}
		}
	}
	return out
}

func promptUser(name string) string {
	fmt.Printf("Please provide %s: ", name)
	out, err := scanner.ReadString('\n')
	if err != nil {
		log.Exitf("error reading line: %v", err)
	}
	out = strings.ReplaceAll(out, "\n", "")
	return out
}
