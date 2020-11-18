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
	"context"
	"flag"
	"time"

	log "github.com/golang/glog"
	gnoiOS "github.com/google/gnxi/gnoi/os"
	"github.com/google/gnxi/utils/credentials"
	"google.golang.org/grpc"
)

var (
	targetAddr    = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	version       = flag.String("version", "", "Version of the OS required when using the activate operation or being installed using the install operation")
	osFile        = flag.String("os", "", "Path to the OS image for the install operation")
	op            = flag.String("op", "", "OS service operation. Can be one of: install, activate, verify")
	timeOut       = flag.Duration("time_out", 5*time.Second, "Timeout for the operation, 5 seconds by default")
	readChunkSize = flag.Uint64("chunk_size", 4000000, "How much of the image to load a time, in bytes. Example: -chunk_size 4000000")

	client *gnoiOS.Client
	ctx    context.Context
	cancel func()
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	opts := credentials.ClientCredentials()
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	client = gnoiOS.NewClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	switch *op {
	case "install":
		install()
	case "activate":
		activate()
	case "verify":
		verify()
	default:
		flag.Usage()
		log.Error("Invalid operation provided. Provide one with -op")
	}
}

// install installs the OS image onto the target.
func install() {
	if *osFile == "" {
		log.Exit("No OS image path provided. Provide one with -os")
	}
	if err := client.Install(ctx, *osFile, *version, *timeOut, *readChunkSize); err != nil {
		log.Exit("Failed Install: ", err)
	}
}

// activate activates the OS version to be used upon next reboot on the target.
func activate() {
	if *version == "" {
		log.Exit("No version provided. Provide one with -version")
	}
	if err := client.Activate(ctx, *version); err != nil {
		log.Exit("Failed Activate: ", err)
	}
}

// verify verifies the version of the OS running on the target.
func verify() {
	version, activationFailMsg, err := client.Verify(ctx)
	if err != nil {
		log.Exit("Failed Verify: ", err)
	}
	if activationFailMsg != "" {
		log.Info("Previous activation fail message:", activationFailMsg)
	}
	log.Info("Running OS version: ", version)
}
