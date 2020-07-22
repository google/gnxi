package orchestrator

import "github.com/google/gnxi/gnxi_tester/config"

// RunTests will take in test name and run each test or all tests.
func RunTests(tests []string) (success string, err error) {
	if len(tests) == 0 {
		configTests := config.GetTests()
		for name := range configTests {
			runTest(name)
		}
	} else {
		for _, name := range tests {
			runTest(name)
		}
	}
	return
}

func runTest(test string) (string, error) {
	return "", nil
}
