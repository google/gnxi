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
	"flag"

	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gnxi_tester",
		Short: "A client tester for the gNxI protocols.",
		Long:  "A client utility that will run each of the client service binaries on a target and validate that the responses are correct.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.CommandLine.Parse([]string{}) // Surpresses logging before flag.Parse error
		},
	}
	cfgPath string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(targetsCmd)
	rootCmd.PersistentFlags().StringVar(&cfgPath, "cfg", "", "Path to the config file.")
}

func initConfig() {
	config.Init(cfgPath)
}

// Execute the root command.
func Execute() error {
	return rootCmd.Execute()
}
