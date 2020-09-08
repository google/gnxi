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
	"os"
	"path"

	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/google/gnxi/gnxi_tester/orchestrator"
	"github.com/spf13/cobra"
)

var (
	wipeCmd = &cobra.Command{
		Use:     "wipe",
		Short:   "Wipe config file and containers...",
		Example: "gnxi_tester wipe [files, containers]",
		Run:     handleWipe,
	}
)

// handleWipe wipes files and containers from the host.
func handleWipe(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		args = []string{"files", "containers"}
	}
	for _, arg := range args {
		switch arg {
		case "files":
			wipeFiles()
			log.Infof("Removed files")
		case "containers":
			tests := config.GetTests()
			names := []string{}
			for test := range tests {
				if test != "provision" {
					names = append(names, test)
				}
			}
			if err := orchestrator.WipeContainers(names); err != nil {
				log.Error(err)
				return
			}
			log.Info("Deleted containers and images")
		}
	}
}

// wipeFiles deletes .gnxi.yml and .gnxi
func wipeFiles() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
	}
	os.Remove(path.Join(home, ".gnxi.yml"))
	os.RemoveAll(path.Join(home, ".gnxi"))
}
