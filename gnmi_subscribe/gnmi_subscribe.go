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
	"errors"
	"flag"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils"
	"github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/xpath"
	"github.com/openconfig/gnmi/proto/gnmi"
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
	defaultRequestTimeout = 10 // This represents a value of 10 seconds and is used as a default RPC request timeout value.
)

var (
	subscribeClient   pb.GNMI_SubscribeClient
	xPathFlags        arrayFlags
	pbPathFlags       arrayFlags
	targetAddr        = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	targetName        = flag.String("target_name", "", "The target name used to verify the hostname returned by TLS handshake")
	connectionTimeout = flag.Duration("timeout", defaultRequestTimeout*time.Second, "The timeout for a request in seconds, 10 seconds by default")
	subscriptionOnce  = flag.Bool("once", false, "If true, the target sends values once off")
	subscriptionPoll  = flag.Bool("poll", false, "If true, the target sends values on request")
	streamOnChange    = flag.Bool("stream_on_change", false, "If true, the target sends updates on change")
	sampleInterval    = flag.Uint64("sample_interval", 0, "If defined, the target sends sample values according to this interval in nano seconds")
	encodingFormat    = flag.String("encoding", "JSON_IETF", "The encoding format used by the target for notifications")
	suppressRedundant = flag.Bool("suppress_redundant", false, "If true, in SAMPLE mode, unchanged values are not sent by the target")
	heartbeatInterval = flag.Uint64("heartbeat_interval", 0, "Specifies maximum allowed period of silence in seconds when surpress redundant is used")
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

	ctx, cancel := context.WithTimeout(context.Background(), *connectionTimeout)
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

	var subscriptionListMode gnmi.SubscriptionList_Mode
	switch {
	case *subscriptionPoll && *subscriptionOnce:
		flag.Usage()
		log.Exitf("Only one of -once and -poll can be set")
	case *subscriptionOnce:
		subscriptionListMode = pb.SubscriptionList_ONCE
	case *subscriptionPoll:
		subscriptionListMode = pb.SubscriptionList_POLL
	default:
		subscriptionListMode = pb.SubscriptionList_STREAM
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
	subscriptions, err := assembleSubscriptions(pbPathList)
	if err != nil {
		log.Exitf("Error assembling subscriptions: %v", err)
	}

	request := &pb.SubscribeRequest{
		Request: &pb.SubscribeRequest_Subscribe{
			Subscribe: &pb.SubscriptionList{
				Encoding:     pb.Encoding(encoding),
				Mode:         subscriptionListMode,
				Subscription: subscriptions,
			},
		},
	}
	utils.PrintProto(request)
	if err := subscribeClient.Send(request); err != nil {
		log.Exitf("Failed to send request: %v", err)
	}
	switch subscriptionListMode {
	case pb.SubscriptionList_STREAM:
		stream()
	case pb.SubscriptionList_POLL:
		poll()
	case pb.SubscriptionList_ONCE:
		if err := once(); err != nil {
			log.Exitf("Error using ONCE mode: %v", err)
		}
	}
}

func stream() {

}

func poll() {

}

func once() error {
	for {
		res, err := subscribeClient.Recv()
		if err != nil {
			return err
		}
		switch res.Response.(type) {
		case *pb.SubscribeResponse_SyncResponse:
			if syncRes := res.GetSyncResponse(); syncRes {
				log.Info("Received all updates")
				return nil
			}
		case *pb.SubscribeResponse_Update:
			utils.LogProto(res)
		default:
			return errors.New("Unexpected response type")
		}
	}
}

func assembleSubscriptions(paths []*pb.Path) ([]*pb.Subscription, error) {
	var subscriptions []*pb.Subscription
	var subscriptionMode gnmi.SubscriptionMode
	switch {
	case *streamOnChange && *sampleInterval != 0:
		return nil, errors.New("Only one of -stream_on_change and -sample_interval can be set")
	case *streamOnChange:
		subscriptionMode = pb.SubscriptionMode_ON_CHANGE
	case *sampleInterval != 0:
		subscriptionMode = pb.SubscriptionMode_SAMPLE
	default:
		subscriptionMode = pb.SubscriptionMode_TARGET_DEFINED
	}
	for _, path := range paths {
		subscription := &pb.Subscription{
			Path:              path,
			Mode:              subscriptionMode,
			SampleInterval:    *sampleInterval,
			SuppressRedundant: *suppressRedundant,
			HeartbeatInterval: *heartbeatInterval,
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}
