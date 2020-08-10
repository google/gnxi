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
	"github.com/google/gnxi/gnxi_tester/web"
	"github.com/spf13/cobra"
)

var (
	webCmd = &cobra.Command{
		Use:     "web",
		Short:   "Run web api",
		Long:    "Run web api for use with web_ui",
		Example: "gnxi_tester web",
		Run:     handleWeb,
	}
	laddr string
)

func init() {
	webCmd.Flags().StringVarP(&laddr, "local_address", "l", "localhost:8888", "The host + port to listen on")
}

// handleWeb will run the web api.
func handleWeb(cmd *cobra.Command, args []string) {
	web.InitRouter(laddr)
}
