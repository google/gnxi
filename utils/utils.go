// Package utils implements utilities for gnxi.
package utils

import (
	"flag"
	"fmt"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
)

var (
	usePretty   = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal")
	logProto    = flag.Bool("log_proto", false, "If true it prints all sent and received PROTO messages")
	printStatus = flag.Bool("print_status", false, "Print status of file transfer.")
)

// PrintProto prints a Proto in a structured way.
func PrintProto(m proto.Message) {
	fmt.Println(FormatProto(m))
}

// PrintStatus prints the percentage transferred if print_status is true
func PrintStatus(transferred, fileSize uint64) {
	if *printStatus {
		fmt.Printf("%d%% transferred\n", transferred/fileSize)
	}
}

// LogProto prints a Proto message only if the log_proto flag is true.
func LogProto(m proto.Message) {
	if *logProto {
		log.Info(FormatProto(m))
	}
}

// FormatProto formats the proto string according to the pretty flag
func FormatProto(m proto.Message) string {
	if *usePretty {
		return pretty.Sprint(m)
	}
	return proto.MarshalTextString(m)
}
