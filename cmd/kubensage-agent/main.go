package main

import (
	"context"
	"google.golang.org/grpc"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
)

func main() {
	// Capture termination signals for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Create a 5-second ticker
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Context to optionally pass to Discover
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Discover the CRI socket
	socket, err := discovery.CriSocketDiscovery()
	if err != nil {
		log.Fatalf("Failed to discover CRI socket: %v", err)
	}

	// Create the gRPC connection
	conn, err := utils.GrpcClientConnection(socket)
	if err != nil {
		log.Fatalf("Failed to connect to CRI socket: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}(conn)

	// Create the runtime client
	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)

	log.Println("Starting kubensage-agent loop, polling every 5s...")

	for {
		select {
		case <-sigCh:
			log.Println("Termination signal received, exiting.")
			return

		case <-ticker.C:
			// Launch the collection in the background
			go func() {
				if err := collectOnce(ctx, runtimeClient); err != nil {
					log.Printf("Error in collectOnce: %v", err)
				}
			}()
		}
	}
}

// collectOnce calls discovery.Discover and logs the results
func collectOnce(ctx context.Context, runtimeClient runtimeapi.RuntimeServiceClient) error {
	podInfos, err := discovery.Discover(ctx, runtimeClient)
	if err != nil {
		return err
	}

	for _, podInfo := range podInfos {
		jsonStr, err := utils.ToJsonString(podInfo)
		if err != nil {
			log.Printf("Error serializing PodInfo: %v", err)
			continue
		}
		log.Printf("PodInfo: %s", jsonStr)
	}
	return nil
}
