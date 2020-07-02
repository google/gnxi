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
	"time"

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

	bindAddr             = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	resetDelay           = flag.Duration("reset_delay", 3*time.Second, "Delay before resetting the service upon factory reset request, 3 seconds by default")
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

// notifyCerts can be called with the number of certs and ca certs installed. It will
// (re)start the gRPC server in encrypted mode if no certs are installed. It will
// (re)start in authenticated mode otherwise.
func notifyCerts(certs, caCerts int) {
	hasCredentials := certs != 0 && caCerts != 0
	if bootstrapping != hasCredentials {
		// Nothing to do, either I am boostrapping and I have no
		// certificates or I am provisioned and I have certificates.
		return
	}
	if bootstrapping {
		log.Info("Found Credentials, setting Provisioned state.")
		grpcServer.GracefulStop()
		grpcServer = gNOIServer.PrepareAuthenticated()
		// Register all gNOI services.
		gNOIServer.Register(grpcServer)
	} else {
		log.Info("No credentials, setting Bootstrapping state.")
		if grpcServer != nil {
			grpcServer.GracefulStop()
		}
		grpcServer = gNOIServer.PrepareEncrypted()
		// Only register the gNOI Cert service for bootstrapping.
		gNOIServer.RegCertificateManagement(grpcServer)
	}
	bootstrapping = !bootstrapping
	go serve()
}

// start creates the new gNOI server.
func start() {
	resetSettings := &reset.Settings{
		ZeroFillUnsupported:  *zeroFillUnsupported,
		FactoryOSUnsupported: *factoryOSUnsupported,
	}
	var err error
	if gNOIServer, err = gnoi.NewServer(nil, nil, resetSettings, notifyReset); err != nil {
		log.Fatal("Failed to create gNOI Server:", err)
	}
	// Registers a caller for whenever the number of installed certificates changes.
	gNOIServer.RegisterCertNotifier(notifyCerts)
	bootstrapping = false
	notifyCerts(0, 0) // Trigger bootstraping mode.
}

// notifyReset is called when the factory reset service requires the server
// to be restarted.
func notifyReset() {
	log.Info("Server factory reset triggered")
	<-time.After(*resetDelay)
	start()
}

func main() {
	flag.Parse()
	start()
	select {} // Loop forever.
}
