package main

import (
	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Exitf("an error occured: %v", err)
	}
}
