package cmd

import "github.com/spf13/cobra"

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run set of tests.",
	Long:    "Run a set of tests from the config file",
	Example: "gnxi_tester run [test_names]",
	RunE:    handleRun,
}

// handleRun will run some or all of the tests.
func handleRun(cmd *cobra.Command, args []string) error {
	return nil
}
