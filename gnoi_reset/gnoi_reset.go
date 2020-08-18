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
	targetAddr = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	rollbackOs = flag.Bool("rollback_os", false, "Target must revert to factory OS")
	zeroFill   = flag.Bool("zero_fill", false, "Target must overwrite persistent storage with zeroes")
	timeOut    = flag.Duration("time_out", 10*time.Second, "Timeout for ResetTarget operation, 10 seconds by default")
)

func main() {
	flag.Parse()

	opts := credentials.ClientCredentials()
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	client := reset.NewClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	if err := client.ResetTarget(ctx, *rollbackOs, *zeroFill); err != nil {
		log.Errorf("Error Resetting Target: %v", err)
	} else {
		log.Infoln("Reset Called Successfully!")
	}
}
