package main

import (
	"context"
	"flag"
	"time"

	log "github.com/golang/glog"
	gnoiOS "github.com/google/gnxi/gnoi/os"
	"github.com/google/gnxi/utils/credentials"
	"google.golang.org/grpc"
)

var (
	targetAddr = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	targetName = flag.String("target_name", "", "The target name used to verify the hostname returned by TLS handshake")
	version    = flag.String("version", "", "Version of the OS required when using the activate operation")
	osFile     = flag.String("os", "", "Path to the OS image for the install operation")
	op         = flag.String("op", "", "OS service operation. Can be one of: install, activate, verify")
	timeOut    = flag.Duration("time_out", 5*time.Second, "Timeout for the operation, 5 seconds by default")

	client *gnoiOS.Client
	ctx    context.Context
	cancel func()
)

func main() {
	flag.Parse()

	if *targetName == "" {
		flag.Usage()
		log.Exit("-target_name must be specified")
	}
	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	client = gnoiOS.NewClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), *timeOut)
	defer cancel()

	switch *op {
	case "install":
		install()
	case "activate":
		activate()
	case "verify":
		verify()
	default:
		log.Error("No operation provided. Provide one with -op")
	}
}

func install() {
	if *osFile == "" {
		log.Error("No OS image path provided. Provide one with -os")
		return
	}
	// TODO: Add Install RPC call
}

func activate() {
	if *version == "" {
		log.Exit("No version provided. Provide one with -version")
	}
	err := client.Activate(ctx, *version)
	if err != nil {
		log.Exit("Failed Activate:", err)
	}
}

func verify() {
	// TODO: Add Verify RPC call
}
