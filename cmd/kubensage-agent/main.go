package main

import (
	"context"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
)

var TickerDuration = time.Second * 5

func main() {
	utils.SetupLogging("kubensage-agent.log")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(TickerDuration)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	socket, err := discovery.CriSocketDiscovery()
	if err != nil {
		log.Fatalf("Failed to discover CRI socket: %v", err)
	}

	grpcConnection := utils.AcquireGrpcConnection(socket)

	runtimeClient := runtimeapi.NewRuntimeServiceClient(grpcConnection)

	log.Println("Starting kubensage-agent loop, polling every 5s...")

	for {
		select {
		case <-sigCh:
			log.Println("Termination signal received, exiting.")
			return

		// Collection section
		case <-ticker.C:
			go func() {
				metrics, err := discovery.GetAllMetrics(ctx, runtimeClient)

				if err != nil {
					log.Printf("Failed to get metrics: %v", err)
				} else {
					log.Println(utils.ToJsonString(metrics))
				}
			}()
		}
	}
}
