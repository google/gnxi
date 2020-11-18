// Package utils implements utilities for gnxi.
package utils

import (
	"flag"
	"fmt"

	"github.com/golang/protobuf/proto"
)

var (
	usePretty    = flag.Bool("pretty", false, "Deprecated, please dont use.")
	logProto     = flag.Bool("log_proto", false, "Deprecated, please dont use.")
	printProgess = flag.Bool("print_progress", false, "Prints progress periodically of file transfer.")
)

// PrintProto prints a Proto in a structured way.
func PrintProto(m proto.Message) {
	fmt.Println(proto.MarshalTextString(m))
}

// PrintProgress prints the percentage transferred if print_status is true
func PrintProgress(progress string) {
	if *printProgess {
		fmt.Println(progress)
	}
}
