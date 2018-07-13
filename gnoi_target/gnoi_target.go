package main

import (
	"flag"
	"net"
	"sync"
	"time"

	"github.com/google/gnxi/gnoi"
	"google.golang.org/grpc"

	log "github.com/golang/glog"
)

var (
	conString = "127.0.0.1:45444"
)

func main() {
	flag.Parse()

	g, err := gnoi.NewServer(nil, nil)
	if err != nil {
		log.Fatal("Failed to create gNOI Server:", err)
	}

	var grpcServer *grpc.Server
	var ready sync.Mutex
	serve := func() {
		ready.Lock()
		defer ready.Unlock()
		listen, err := net.Listen("tcp", conString)
		if err != nil {
			log.Fatal("Failed to listen:", err)
		}
		log.Info("Starting gNOI server.")
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal("Failed to serve:", err)
		}
	}

	bootstrapping := false
	for {
		if bootstrapping != !g.HasCredentials() {
			if bootstrapping {
				log.Info("Found Credentials, setting Provisioned state.")
				grpcServer.GracefulStop()
				grpcServer = g.PrepareAuthenticated()
				g.RegCertificateManagement(grpcServer)
			} else {
				log.Info("No credentials, setting Bootstrapping state.")
				if grpcServer != nil {
					grpcServer.GracefulStop()
				}
				grpcServer = g.PrepareEncrypted()
				g.RegCertificateManagement(grpcServer)
			}
			bootstrapping = !bootstrapping
			go serve()
		}
		time.Sleep(time.Second)
	}
}
