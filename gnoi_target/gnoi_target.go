package main

import (
	"flag"
	"net"

	"github.com/google/gnxi/gnoi"

	log "github.com/golang/glog"
)

var (
	conString = "127.0.0.1:45444"
)

func main() {
	flag.Parse()

	log.Info("Starting gNOI server.")
	g, err := gnoi.NewServer(nil, nil)
	if err != nil {
		log.Fatal("Failed to create gNOI Server:", err)
	}

	listen, err := net.Listen("tcp", conString)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := g.PrepareEncrypted()
	g.RegCertificateManagement(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal("Failed to serve:", err)
	}

	log.Info("Graceful exit.")
}

// NewBootstrappingServer returns a new BootstrappingServer.
// func NewBootstrappingServer(privateKey crypto.PrivateKey, defaultCertificate *tls.Certificate) (*BootstrappingServer, error) {
// 	s, err := NewServer(privateKey, defaultCertificate)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &BootstrappingServer{Server: s, bootstrapping: true}, nil
// }
//
// // Prepare does blah.
// func (bs *BootstrappingServer) Prepare(l net.Listener) (func() error, func()) {
// 	bs.mu.Lock()
// 	defer bs.mu.Unlock()
// 	bs.grpcServer = bs.Server.PrepareEncrypted()
//
// 	go func() {
// 		for {
// 			if bs.bootstrapping != bs.Server.certManager.Empty() {
// 				bs.mu.Lock()
// 				if bs.bootstrapping {
// 					bs.grpcServer.GracefulStop()
// 					bs.grpcServer = bs.Server.PrepareAuthenticated()
// 				} else {
// 					bs.grpcServer.GracefulStop()
// 					bs.grpcServer = bs.Server.PrepareEncrypted()
// 				}
// 				bs.bootstrapping = !bs.bootstrapping
// 				bs.mu.Unlock()
// 				time.Sleep(time.Second)
// 			}
// 		}
// 	}()
//
// 	serve := func() error {
// 		return bs.grpcServer.Serve(l)
// 	}
//
// 	stop := func() {
// 		bs.mu.Lock()
// 		defer bs.mu.Unlock()
// 		bs.grpcServer.GracefulStop()
// 	}
//
// 	return serve, stop
// }
