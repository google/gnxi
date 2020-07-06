package main

import (
	"flag"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/mockos"
)

var (
	file      = flag.String("file", "", "The name and path of the OS file")
	version   = flag.String("version", "1.0a", "The version of the OS package")
	size      = flag.String("size", "", "The size of the OS package's data, e.g 10M")
	supported = flag.Bool("supported", true, "Determines if the OS package is supported by the mock target")
)

func main() {
	flag.Parse()

	if *file == "" || *size == "" {
		flag.Usage()
		log.Exit("-file and -size must be specified")
	}

	if err := mockos.GenerateOS(*file, *version, *size, *supported); err != nil {
		log.Exitf("Error Generating OS: %v", err)
	}
	log.Info("OS Generated Successfully")
}
