package main

import (
	"flag"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/credentials"
	"google.golang.org/grpc"
)

var (
	targetAddr = flag.String("target_addr", ":9339", "The target address in the format of host:port")
	targetName = flag.String("target_name", "", "The target name used to verify the hostname returned by TLS handshake")
	version    = flag.String("version", "", "Version of the OS required when using the activate operation")
	osFile     = flag.String("os", "", "Path to the OS image for the install operation")
	op         = flag.String("op", "", "OS service operation. Can be one of: install, activate, verify")
)

func main() {
	flag.Parse()

	if *targetName == "" || *targetAddr == "" {
		flag.Usage()
		log.Exit("-target_name and -target_addr must be specified")
	}
	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Exitf("Dialing to %s failed: %v", *targetAddr, err)
	}
	defer conn.Close()

	// TODO: Setup client

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
}

func activate() {
	if *version == "" {
		log.Error("No version provided. Provide one with -version")
		return
	}
}

func verify() {

}
