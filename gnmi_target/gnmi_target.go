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

// Package gnmi_target implements an in-memory gnmi target for device config.
package main

// Typical usage:
// go run gnmi_target.go -bind :10161 \
//		-config openconfig-openflow.json \
//		-ca ca.pem -cert target_cert.pem -key target_key.pem \
//		-username foo -password bar

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/google/gnxi/credentials"
	"github.com/google/gnxi/utils/target/gnmi/model_data"
	"github.com/google/gnxi/utils/target/gnmi/model_data/gostruct"
	"github.com/google/gnxi/utils/target/gnmi/server"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

type target struct {
	*server.Server
}

func new(model *server.Model, config []byte) (*target, error) {
	s, err := server.NewServer(model, config, nil)
	if err != nil {
		return nil, err
	}
	return &target{Server: s}, nil
}

// Get overrides the Get func of server.Server to provide user auth.
func (t *target) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Get request: %v", msg)
	return t.Server.Get(ctx, req)
}

// Set overrides the Set func of server.Server to provide user auth.
func (t *target) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Get request: %v", msg)
	return t.Server.Set(ctx, req)
}

var (
	bindAddr   = flag.String("bind", ":10161", "Bind to address:port or just :port.")
	configFile = flag.String("config", "", "IETF JSON file for target startup config")
)

func main() {
	model := server.NewModel(model_data.ModelData,
		reflect.TypeOf((*gostruct.Device)(nil)),
		gostruct.SchemaTree["Device"],
		gostruct.Unmarshal)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Exitf("error in reading config file: %v", err)
		}
	}
	t, err := new(model, configData)
	if err != nil {
		log.Exitf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, t)
	reflection.Register(g)

	log.Infof("starting to listen on", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Exitf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Exitf("failed to serve: %v", err)
	}
}
