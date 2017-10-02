// Package utils implements utilities for gnxi.
package utils

import (
	"flag"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
)

var (
	usePretty = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal")
)

// PrintProto prints a Proto in a structured way.
func PrintProto(m proto.Message) {
	if *usePretty {
		pretty.Print(m)
		return
	}
	fmt.Println(proto.MarshalTextString(m))
}
