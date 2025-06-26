package main

import (
	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"log"
)

func main() {
	err := discovery.Discover()

	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}
