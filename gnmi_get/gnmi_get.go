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

// Binary gnmi_get performs a get request against a gnmi target.
package main

// Typical usage:
// go run gnmi_get.go \
//		-xpath "/system/openflow/agent/config/datapath-id" \
//		-xpath "/system/openflow/agent/config/backoff-interval" \
//		-target_addr localhost:10161 -target_name www.example.com \
//		-ca ca.pem -cert client_cert.pem -key client_key.pem \
//		-username foo -password bar

import (
	"flag"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/gnxi/credentials"
	"github.com/google/gnxi/utils/xpath"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	xPathFlags  arrayFlags
	pbPathFlags arrayFlags
	targetAddr  = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName  = flag.String("target_name", "www.example.com", "The target name use to verify the hostname returned by TLS handshake")
	timeOut     = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default.")
	usePretty   = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal.")
)

func display(m proto.Message) {
	if *usePretty {
		pretty.Print(m)
		return
	}
	fmt.Println(proto.MarshalTextString(m))
}

func main() {
	flag.Var(&xPathFlags, "xpath", "xpath of the config node to be fetched")
	flag.Var(&pbPathFlags, "pbpath", "protobuf format path of the config node to be fetched")
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

	var pbPathList []*pb.Path
	for _, xPath := range xPathFlags {
		pbPath, err := xpath.ToGNMIPath(xPath)
		if err != nil {
			log.Exitf("error in parsing xpath %q to gnmi path", xPath)
		}
		pbPathList = append(pbPathList, pbPath)
	}
	for _, textPbPath := range pbPathFlags {
		var pbPath pb.Path
		if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
			log.Exitf("error in unmarshaling %q to gnmi Path", textPbPath)
		}
		pbPathList = append(pbPathList, &pbPath)
	}

	getRequest := &pb.GetRequest{
		Path:      pbPathList,
		Encoding:  pb.Encoding_JSON_IETF,
		UseModels: capResponse.GetSupportedModels(),
	}
	getResponse, err := cli.Get(ctx, getRequest)
	if err != nil {
		log.Exitf("fetch config from the path failed: %v", err)
	}

	fmt.Println("== getRequest:")
	display(getRequest)

	fmt.Println("== getResponse:")
	display(getResponse)
}
