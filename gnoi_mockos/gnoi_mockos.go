/* Copyright 2020 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
