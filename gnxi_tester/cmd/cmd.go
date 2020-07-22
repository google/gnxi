package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "gnxi_tester",
		Short: "A client tester for the gNxI protocols.",
		Long:  "A client utility that will run each of the client service binaries on a target and validate that the responses are correct.",
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
}

// Execute the root command.
func Execute() error {
	return rootCmd.Execute()
}
