// Package utils implements utilities for gnxi.
package utils

import (
	"flag"
)

var (
	usePretty = flag.Bool("pretty", false, "Deprecated, please dont use.")
	logProto  = flag.Bool("log_proto", false, "Deprecated, please dont use.")
)
