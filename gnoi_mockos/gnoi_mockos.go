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
	"path/filepath"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/mockos"
)

var (
	file                  = flag.String("file", "", "The name and path of the OS file")
	version               = flag.String("version", "", "The version of the OS package")
	size                  = flag.String("size", "1M", "The size of the OS package's data, e.g 10M")
	incompatible          = flag.Bool("incompatible", false, "If true, the os package is valid but incompatible with the target")
	activationFailMessage = flag.String("activation_fail_message", "", "If set, then the OS will fail to activate")
)

func main() {
	flag.Parse()

	if *file == "" || *version == "" {
		flag.Usage()
		log.Exit("-file and -version must be specified")
	}

	if err := mockos.GenerateOS(*file, *version, *size, *activationFailMessage, *incompatible); err != nil {
		log.Exitf("Error Generating OS: %v", err)
	}
	path, err := filepath.Abs(*file)
	if err != nil {
		log.Exit(err)
	}
	log.Infof("OS Generated Successfully: %s", path)
}
