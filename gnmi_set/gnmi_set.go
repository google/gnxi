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

// Binary gnmi_set sets (replaces) the target config with the specified config file.
package main

// Typical usage:
// go run gnmi_set.go -config openconfig-openflow.json \
//		-target_addr localhost:10161 -target_name www.example.com \
//		-ca ca.crt -cert client.crt -key client.key \
//		-username foo -password bar

import (
	"flag"
	"io/ioutil"
	"time"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/gnxi/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var (
	targetAddr = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName = flag.String("target_name", "www.example.com", "The target name use to verify the hostname returned by TLS handshake")
	timeOut    = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default")
	configFile = flag.String("config", "", "IETF JSON config file to replace the current config of the target")
)

func main() {
	flag.Parse()
	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %q failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	jsonConfig, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Exitf("error in reading config file: %v", err)
	}
	pbValConfig := &pb.TypedValue{
		Value: &pb.TypedValue_JsonIetfVal{
			JsonIetfVal: jsonConfig,
		},
	}
	req := &pb.SetRequest{
		Replace: []*pb.Update{
			&pb.Update{
				Path: &pb.Path{},
				Val:  pbValConfig,
			},
		},
	}

	cli := pb.NewGNMIClient(conn)
	if _, err := cli.Set(context.Background(), req); err != nil {
		log.Exitf("set config to the target failed: %v", err)
	}
}
