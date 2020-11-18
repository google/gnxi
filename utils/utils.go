// Package utils implements utilities for gnxi.
package utils

import (
	"flag"
	"fmt"
)

var (
	usePretty    = flag.Bool("pretty", false, "Deprecated, please dont use.")
	logProto     = flag.Bool("log_proto", false, "Deprecated, please dont use.")
	printProgess = flag.Bool("print_progress", false, "Prints progress periodically of file transfer.")
)

// PrintProgress prints the percentage transferred if print_status is true
func PrintProgress(progress string) {
	if *printProgess {
		fmt.Println(progress)
	}
}
