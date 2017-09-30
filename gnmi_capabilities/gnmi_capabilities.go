/* Copyright 2017 Google Inc.

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

// Binary gnmi_capabilities performs a capabilities request to a gNMI target.
package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var (
	targetAddr = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName = flag.String("target_name", "hostname.com", "The target name use to verify the hostname returned by TLS handshake")
	timeOut    = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default")
	usePretty  = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal")
)

func main() {
	flag.Parse()

	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %q failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	cli := pb.NewGNMIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	capResponse, err := cli.Capabilities(ctx, &pb.CapabilityRequest{})
	if err != nil {
		log.Exitf("error in getting capabilities: %v", err)
	}

	fmt.Println("== capabilitiesResponse:")
	utils.PrintProto(capResponse)
}
