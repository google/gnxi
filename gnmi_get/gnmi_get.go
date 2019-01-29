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

// Binary gnmi_get performs a get request against a gNMI target.
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/credentials"
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

func parseModelData(s *string) (*pb.ModelData, error) {
	pbModelData := new(pb.ModelData)
	modelDataVars := strings.Split(*s, ",")
	if len(modelDataVars) != 3 {
		return pbModelData, fmt.Errorf("Unable to parse string")
	}
	pbModelData = &pb.ModelData{
		Name:         modelDataVars[0],
		Organization: modelDataVars[1],
		Version:      modelDataVars[2],
	}
	return pbModelData, nil
}

var (
	xPathFlags       arrayFlags
	pbPathFlags      arrayFlags
	pbModelDataFlags arrayFlags
	targetAddr       = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName       = flag.String("target_name", "hostname.com", "The target name use to verify the hostname returned by TLS handshake")
	timeOut          = flag.Duration("time_out", 10*time.Second, "Timeout for the Get request, 10 seconds by default")
	encodingName     = flag.String("encoding", "JSON_IETF", "value encoding format to be used")
)

func main() {
	flag.Var(&xPathFlags, "xpath", "xpath of the config node to be fetched")
	flag.Var(&pbPathFlags, "pbpath", "protobuf format path of the config node to be fetched")
	flag.Var(&pbModelDataFlags, "model_data", "Data models to be used by the target in the format of 'name,organization,version'")
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

	encoding, ok := pb.Encoding_value[*encodingName]
	if !ok {
		var gnmiEncodingList []string
		for _, name := range pb.Encoding_name {
			gnmiEncodingList = append(gnmiEncodingList, name)
		}
		log.Exitf("Supported encodings: %s", strings.Join(gnmiEncodingList, ", "))
	}

	var pbPathList []*pb.Path
	var pbModelDataList []*pb.ModelData
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
	for _, textPbModelData := range pbModelDataFlags {
		pbModelData, err := parseModelData(&textPbModelData)
		if err == nil {
			pbModelDataList = append(pbModelDataList, pbModelData)
		}
	}

	getRequest := &pb.GetRequest{
		Encoding:  pb.Encoding(encoding),
		Path:      pbPathList,
		UseModels: pbModelDataList,
	}

	fmt.Println("== getRequest:")
	utils.PrintProto(getRequest)

	getResponse, err := cli.Get(ctx, getRequest)
	if err != nil {
		log.Exitf("Get failed: %v", err)
	}

	fmt.Println("== getResponse:")
	utils.PrintProto(getResponse)
}
