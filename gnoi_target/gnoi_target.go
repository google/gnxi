/* Copyright 2018 Google Inc.

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

// Binary implements a gNOI Target with a Certificate Management service.
package main

import (
	"flag"
	"net"
	"sync"

	"github.com/google/gnxi/gnoi"
	"github.com/google/gnxi/gnoi/reset"
	"google.golang.org/grpc"

	log "github.com/golang/glog"
)

var (
	gNOIServer    *gnoi.Server
	grpcServer    *grpc.Server
	muServe       sync.Mutex
	bootstrapping bool
	resetSettings *reset.Settings

	bindAddr             = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	zeroFillUnsupported  = flag.Bool("zero_fill_unsupported", false, "Make the target not support zero filling storage")
	factoryOSUnsupported = flag.Bool("reset_unsupported", false, "Make the target not support factory resetting OS")
)

// serve binds to an address and starts serving a gRPCServer.
func serve() {
	muServe.Lock()
	defer muServe.Unlock()
	listen, err := net.Listen("tcp", *bindAddr)
	defer listen.Close()
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	log.Info("Starting gNOI server.")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// notifyReset will be called when the factory reset service requires the server
// to be restarted.
func notifyReset() {
	var err error
	if gNOIServer, err = gnoi.NewServer(nil, nil, resetSettings); err != nil {
		log.Fatal("Failed to create gNOI Server:", err)
	}
	gNOIServer.Register(grpcServer)
}

// notifyCerts can be called with the number of certs and ca certs installed. It will
// (re)start the gRPC server in encrypted mode if no certs are installed. It will
// (re)start in authenticated mode otherwise.
func notifyCerts(certs, caCerts int) {
	hasCredentials := certs != 0 && caCerts != 0
	if bootstrapping != !hasCredentials {
		if bootstrapping {
			log.Info("Found Credentials, setting Provisioned state.")
			grpcServer.GracefulStop()
			grpcServer = gNOIServer.PrepareAuthenticated()
			gNOIServer.Register(grpcServer)
		} else {
			log.Info("No credentials, setting Bootstrapping state.")
			if grpcServer != nil {
				grpcServer.GracefulStop()
			}
			grpcServer = gNOIServer.PrepareEncrypted()
			// Register all gNOI services.
			gNOIServer.Register(grpcServer)
		}
		bootstrapping = !bootstrapping
		go serve()
	}
}

func main() {
	flag.Parse()

	resetSettings = &reset.Settings{
		ZeroFillUnsupported:  *zeroFillUnsupported,
		FactoryOSUnsupported: *factoryOSUnsupported,
	}

	// Registers a caller for whenever the number of installed certificates changes.
	gNOIServer.RegisterCertNotifier(notifyCerts)
	gNOIServer.RegisterResetNotifier(notifyReset)
	notifyCerts(0, 0) // Trigger bootstraping mode.
	select {}         // Loop forever.
}
