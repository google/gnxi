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

// Binary gnmi_set performs a set request against a gNMI target with the specified config file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/credentials"

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
	jsonUpdate  arrayFlags
	jsonReplace arrayFlags
	targetAddr  = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName  = flag.String("target_name", "hostname.com", "The target name use to verify the hostname returned by TLS handshake")
	timeOut     = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default")
)

func jsonToUpdate(jsonFiles arrayFlags, pbUpdates *[]*pb.Update) error {
	for _, jsonFile := range jsonFiles {
		jsonConfig, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("error reading json file: %v", err)
		}
		pbValConfig := &pb.TypedValue{
			Value: &pb.TypedValue_JsonIetfVal{
				JsonIetfVal: jsonConfig,
			},
		}
		update := &pb.Update{
			Path: &pb.Path{},
			Val:  pbValConfig,
		}
		*pbUpdates = append(*pbUpdates, update)
	}
	return nil
}

func main() {
	flag.Var(&jsonUpdate, "json_update", "IETF JSON files to use as update")
	flag.Var(&jsonReplace, "json_replace", "IETF JSON files to use as replace")
	flag.Parse()

	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %q failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	setRequest := &pb.SetRequest{
		Replace: []*pb.Update{},
		Update:  []*pb.Update{},
	}

	if err := jsonToUpdate(jsonUpdate, &setRequest.Update); err != nil {
		log.Exit(err)
	}

	if err := jsonToUpdate(jsonReplace, &setRequest.Replace); err != nil {
		log.Exit(err)
	}

	fmt.Println("== getRequest:")
	utils.PrintProto(setRequest)

	cli := pb.NewGNMIClient(conn)
	setResponse, err := cli.Set(context.Background(), setRequest)
	if err != nil {
		log.Exitf("Set failed: %v", err)
	}

	fmt.Println("== getResponse:")
	utils.PrintProto(setResponse)
}
