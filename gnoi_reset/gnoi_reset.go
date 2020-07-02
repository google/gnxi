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

// Binary implements a gNOI Reset.

package main

import (
	"context"
	"flag"
	"time"

	"google.golang.org/grpc"

	log "github.com/golang/glog"

	"github.com/google/gnxi/gnoi/reset"
	"github.com/google/gnxi/utils/credentials"
)

var (
	targetAddr = flag.String("target_addr", "localhost:9399", "The target address in the format of host:port")
	targetName = flag.String("target_name", "hostname.com", "The target name used to verify the hostname returned by TLS handshake")
	rollbackOs = flag.Bool("rollback_os", false, "Indicate if target should attempt to revert to factory os")
	zeroFill   = flag.Bool("zero_fill", false, "Indicate if target should attempt to overwrite persistent storage with zeroes")
)

func main() {
	flag.Parse()

	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	client := reset.NewClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.ResetTarget(ctx, *rollbackOs, *zeroFill); err != nil {
		log.Errorf("Error Resetting Target: %v", err)
	} else {
		log.Infoln("Reset Called Successfully!")
	}
}
