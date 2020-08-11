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
	"fmt"
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
	defaultRequestTimeout = 10 * time.Second // This represents a value of 10 seconds and is used as a default RPC request timeout value.
)

var (
	xPathFlags        arrayFlags
	pbPathFlags       arrayFlags
	targetAddr        = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	targetName        = flag.String("target_name", "", "The target name used to verify the hostname returned by TLS handshake")
	connectionTimeout = flag.Duration("timeout", defaultRequestTimeout, "The timeout for a request in seconds, 10 seconds by default, e.g 10s")
	subscriptionOnce  = flag.Bool("once", false, "If true, the target sends values once off")
	subscriptionPoll  = flag.Bool("poll", false, "If true, the target sends values on request")
	streamOnChange    = flag.Bool("stream_on_change", false, "If true, the target sends updates on change")
	sampleInterval    = flag.Uint64("sample_interval", 0, "If defined, the target sends sample values according to this interval in nano seconds")
	encodingFormat    = flag.String("encoding", "JSON_IETF", "The encoding format used by the target for notifications")
	suppressRedundant = flag.Bool("suppress_redundant", false, "If true, in SAMPLE mode, unchanged values are not sent by the target")
	heartbeatInterval = flag.Uint64("heartbeat_interval", 0, "Specifies maximum allowed period of silence in seconds when surpress redundant is used")
	updatesOnly       = flag.Bool("updates_only", false, "If true, the target only transmits updates to the subscribed paths")
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

	subscribeClient, err := client.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Error creating GNMI_SubscribeClient: %v", err)
	}

	encoding, err := parseEncoding(*encodingFormat)
	if err != nil {
		log.Exitf("Error parsing encoding: %v", err)
	}

	subscriptionListMode, err := subscriptionMode(*subscriptionPoll, *subscriptionOnce)
	if err != nil {
		flag.Usage()
		log.Exit(err)
	}

	pbPathList, err := parsePaths(xPathFlags, pbPathFlags)
	if err != nil {
		log.Exitf("Error parsing paths: %v", err)
	}

	subscriptions, err := assembleSubscriptions(*streamOnChange, *sampleInterval, pbPathList)
	if err != nil {
		log.Exitf("Error assembling subscriptions: %v", err)
	}

	request := &pb.SubscribeRequest{
		Request: &pb.SubscribeRequest_Subscribe{
			Subscribe: &pb.SubscriptionList{
				Encoding:     pb.Encoding(encoding),
				Mode:         subscriptionListMode,
				Subscription: subscriptions,
				UpdatesOnly:  *updatesOnly,
			},
		},
	}
	utils.PrintProto(request)
	if err := subscribeClient.Send(request); err != nil {
		log.Exitf("Failed to send request: %v", err)
	}

	switch subscriptionListMode {
	case pb.SubscriptionList_STREAM:
		if err := stream(subscribeClient); err != nil {
			log.Exitf("Error using STREAM mode: %v", err)
		}
	case pb.SubscriptionList_POLL:
		if err := poll(subscribeClient); err != nil {
			log.Exitf("Error using POLL mode: %v", err)
		}
	case pb.SubscriptionList_ONCE:
		if err := once(subscribeClient); err != nil {
			log.Exitf("Error using ONCE mode: %v", err)
		}
	}
}

func stream(subscribeClient gnmi.GNMI_SubscribeClient) error {
	for {
		res, err := subscribeClient.Recv()
		if err != nil {
			return err
		}
		switch res.Response.(type) {
		case *pb.SubscribeResponse_SyncResponse:
			log.Info("SyncResponse received")
		case *pb.SubscribeResponse_Update:
			utils.PrintProto(res)
		default:
			return errors.New("Unexpected response type")
		}
	}
}

func poll(subscribeClient gnmi.GNMI_SubscribeClient) error {
	ready := make(chan bool, 1)
	ready <- true
	pollRequest := &pb.SubscribeRequest{Request: &pb.SubscribeRequest_Poll{}}
	if *updatesOnly {
		res, err := subscribeClient.Recv()
		if err != nil {
			return err
		}
		if syncRes := res.GetSyncResponse(); !syncRes {
			return errors.New("Failed to receive SyncResponse")
		}
		log.Info("SyncResponse received")
	}
	for {
		select {
		case <-ready:
			log.Info("Press enter to poll")
			fmt.Scanln()
			subscribeClient.Send(pollRequest)
			utils.LogProto(pollRequest)
		default:
			res, err := subscribeClient.Recv()
			if err != nil {
				return err
			}
			switch res.Response.(type) {
			case *pb.SubscribeResponse_SyncResponse:
				log.Info("SyncResponse received")
				ready <- true
			case *pb.SubscribeResponse_Update:
				utils.PrintProto(res)
			default:
				return errors.New("Unknown response type")
			}
		}
	}

}

func once(subscribeClient gnmi.GNMI_SubscribeClient) error {
	for {
		res, err := subscribeClient.Recv()
		if err != nil {
			return err
		}
		switch res.Response.(type) {
		case *pb.SubscribeResponse_SyncResponse:
			log.Info("SyncResponse received")
			return nil
		case *pb.SubscribeResponse_Update:
			utils.PrintProto(res)
		default:
			return errors.New("Unexpected response type")
		}
	}
}

func assembleSubscriptions(streamOnChange bool, sampleInterval uint64, paths []*pb.Path) ([]*pb.Subscription, error) {
	var subscriptions []*pb.Subscription
	var subscriptionMode gnmi.SubscriptionMode
	switch {
	case streamOnChange && sampleInterval != 0:
		return nil, errors.New("Only one of -stream_on_change and -sample_interval can be set")
	case streamOnChange:
		subscriptionMode = pb.SubscriptionMode_ON_CHANGE
	case sampleInterval != 0:
		subscriptionMode = pb.SubscriptionMode_SAMPLE
	default:
		subscriptionMode = pb.SubscriptionMode_TARGET_DEFINED
	}
	for _, path := range paths {
		subscription := &pb.Subscription{
			Path:              path,
			Mode:              subscriptionMode,
			SampleInterval:    sampleInterval,
			SuppressRedundant: *suppressRedundant,
			HeartbeatInterval: *heartbeatInterval,
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func subscriptionMode(subscriptionPoll, subscriptionOnce bool) (gnmi.SubscriptionList_Mode, error) {
	switch {
	case subscriptionPoll && subscriptionOnce:
		return 0, errors.New("Only one of -once and -poll can be set")
	case subscriptionOnce:
		return pb.SubscriptionList_ONCE, nil
	case subscriptionPoll:
		return pb.SubscriptionList_POLL, nil
	default:
		return pb.SubscriptionList_STREAM, nil
	}
}

func parsePaths(xPathFlags, pbPathFlags arrayFlags) ([]*pb.Path, error) {
	var pbPathList []*pb.Path
	for _, xPath := range xPathFlags {
		pbPath, err := xpath.ToGNMIPath(xPath)
		if err != nil {
			return nil, fmt.Errorf("error in parsing xpath %q to gnmi path", xPath)
		}
		pbPathList = append(pbPathList, pbPath)
	}
	for _, textPbPath := range pbPathFlags {
		var pbPath pb.Path
		if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
			return nil, fmt.Errorf("error in unmarshaling %q to gnmi Path", textPbPath)
		}
		pbPathList = append(pbPathList, &pbPath)
	}
	return pbPathList, nil
}

func parseEncoding(encodingFormat string) (gnmi.Encoding, error) {
	encoding, ok := pb.Encoding_value[encodingFormat]
	if !ok {
		var encodingList []string
		for _, name := range pb.Encoding_name {
			encodingList = append(encodingList, name)
		}
		return -1, fmt.Errorf("Supported encodings: %s", strings.Join(encodingList, ", "))
	}
	return pb.Encoding(encoding), nil
}
