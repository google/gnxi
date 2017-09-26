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

// Binary gnmi_get performs a get request against a gNMI Target.
package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/samribeiro/gnmi/client"
	"github.com/samribeiro/gnmi/credentials"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	targetAddress = flag.String("target_address", "localhost:32123", "The target address:port.")
	targetName    = flag.String("target_name", "", "Will use this hostname to verify server certificate during TLS handshake.")
	timeOut       = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default.")
	query         = flag.String("query", "", "XPath query or queries. Example: system/openflow/controllers/controller[main]/connections/connection[0]/state/address")
	usePretty     = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal.")
)

func display(m proto.Message) {
	if *usePretty {
		pretty.Print(m)
		return
	}
	fmt.Println(proto.MarshalTextString(m))
}

func main() {
	flag.Parse()

	if *query == "" {
		log.Exit("-query must be set")
	}
	queries := client.ParseQuery(*query)
	getRequest, err := client.ToGetRequest(queries)
	if err != nil {
		log.Exitf("error generating GetRequest: %v", err)
	}

	conn, err := grpc.Dial(*targetAddress, credentials.ClientCredentials(*targetName)...)
	if err != nil {
		log.Exitf("did not connect: %v", err)
	}
	defer conn.Close()
	c := gnmi.NewGNMIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	getResponse, err := c.Get(ctx, getRequest)
	if err != nil {
		log.Exitf("could not get: %v", err)
	}

	fmt.Println("== getRequest:")
	display(getRequest)

	fmt.Println("== getResponse:")
	display(getResponse)
}
