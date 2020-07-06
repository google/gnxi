package main

import (
	"flag"
	"fmt"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/osgenerator"
)

var (
	filename  = flag.String("filename", "", "The name of the OS file")
	version   = flag.String("version", "1.0a", "The version of the OS package")
	size      = flag.String("size", "", "The size of the OS package's data, e.g 10M")
	supported = flag.Bool("supported", true, "Specifies if the OS package is supported by the mock target")
)

func main() {
	flag.Parse()

	if *filename == "" || *size == "" {
		flag.Usage()
		log.Exit("-filename and -size must be specified")
	}

	var filesize int
	var unit rune
	if _, err := fmt.Sscanf(*size, "%d%c", &filesize, &unit); err != nil {
		flag.Usage()
		log.Exitf("Invalid Size Specified")
	}
	if err := osgenerator.GenerateOS(*filename, *version, unit, filesize, *supported); err != nil {
		log.Exitf("Error Generating OS: %v", err)
	}
	log.Info("OS Generated Successfully!")
}
