package main

import (
	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"log"
)

func main() {
	podInfos, err := discovery.Discover()

	for _, podInfo := range podInfos {
		jsonStr, err := utils.ToJsonString(podInfo)

		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		log.Printf("PodInfo: %s", jsonStr)
	}

	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}
