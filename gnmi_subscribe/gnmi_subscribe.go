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
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/xpath"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

const (
	defaultTimeout = 10
)

var (
	subscribeClient    pb.GNMI_SubscribeClient
	xPathFlags         arrayFlags
	pbPathFlags        arrayFlags
	targetAddr         = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	targetName         = flag.String("target_name", "", "The target name used to verify the hostname returned by TLS handshake")
	connectionTimeout  = flag.Duration("timeout", defaultTimeout, "The timeout for a request, 10 seconds by default")
	valueStreamingMode = flag.String("value_mode", "STREAM", "The mode of streaming values the target should use use, STREAM by default")
	subscriptionMode   = flag.String("subscription_mode", "TARGET_DEFINED", "The subscription mode the target should use, TARGET_DEFINED by default")
	encodingFormat     = flag.String("encoding", "JSON_IETF", "The encoding format used by the target for notifications")
	allowAggregation   = flag.Bool("allow_aggregation", false, "If true, the elements which are marked eligible are aggregated")
)

func main() {
	flag.Var(&xPathFlags, "xpath", "xpath of the config node to be fetched")
	flag.Var(&pbPathFlags, "pbpath", "protobuf format path of the config node to be fetched")
	flag.Parse()

	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Fatalf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	client := pb.NewGNMIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *connectionTimeout*time.Second)
	defer cancel()

	subscribeClient, err = client.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Error creating GNMI_SubscribeClient: %v", err)
	}

	encoding, ok := pb.Encoding_value[*encodingFormat]
	if !ok {
		var encodingList []string
		for _, name := range pb.Encoding_name {
			encodingList = append(encodingList, name)
		}
		log.Exitf("Supported encodings: %s", strings.Join(encodingList, ", "))
	}

	subscriptionListModeValue, ok := pb.SubscriptionList_Mode_value[*valueStreamingMode]
	if !ok {
		var subscriptionListModes []string
		for _, name := range pb.SubscriptionList_Mode_name {
			subscriptionListModes = append(subscriptionListModes, name)
		}
		log.Exitf("Supported subscription_types: %s", strings.Join(subscriptionListModes, ", "))
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
	subscriptions := assembleSubscriptions(pbPathList)
	if subscriptions == nil {
		log.Exitf("No paths specified!")
	}

	request := &pb.SubscribeRequest{
		Request: &pb.SubscribeRequest_Subscribe{
			Subscribe: &pb.SubscriptionList{
				AllowAggregation: *allowAggregation,
				Encoding:         pb.Encoding(encoding),
				Mode:             pb.SubscriptionList_Mode(subscriptionListModeValue),
				Subscription:     subscriptions,
			},
		},
	}
	utils.PrintProto(request)
	if err := subscribeClient.Send(request); err != nil {
		log.Exitf("Failed to send request: %v", err)
	}
	switch *valueStreamingMode {
	case "STREAM":
		stream()
	case "POLL":
		poll()
	case "ONCE":
		once()
	}
}

func stream() {

}

func poll() {

}

func once() {

}

func assembleSubscriptions(paths []*pb.Path) []*pb.Subscription {
	var subscriptions []*pb.Subscription
	subscriptionModeValue, ok := pb.SubscriptionMode_value[*subscriptionMode]
	if !ok {
		var subscriptionModes []string
		for _, name := range pb.SubscriptionMode_name {
			subscriptionModes = append(subscriptionModes, name)
		}
		log.Exitf("Supported subscription_types: %s", strings.Join(subscriptionModes, ", "))
	}
	for _, path := range paths {
		subscription := &pb.Subscription{
			Path: path,
			Mode: pb.SubscriptionMode(subscriptionModeValue),
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions
}
