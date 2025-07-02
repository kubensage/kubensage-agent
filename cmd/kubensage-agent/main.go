package main

import (
	"context"
	"github.com/kubensage/kubensage-agent/pkg/model"
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
	// Capture termination signals (e.g., SIGINT, SIGTERM) for clean shutdown of the application
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Create a ticker that triggers every 5 seconds to periodically collect pod information
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Create a context that will be passed to Discover function for pod collection
	// It can be cancelled to stop ongoing operations if needed
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Discover the CRI (Container Runtime Interface) socket to connect to the container runtime
	socket, err := discovery.CriSocketDiscovery()
	if err != nil {
		log.Fatalf("Failed to discover CRI socket: %v", err)
	}

	// Establish a gRPC connection to the discovered CRI socket
	conn, err := utils.GrpcClientConnection(socket)
	if err != nil {
		log.Fatalf("Failed to connect to CRI socket: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		// Close the gRPC connection upon termination
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}(conn)

	// Create the runtime client using the gRPC connection
	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)

	// Log the start of the polling loop
	log.Println("Starting kubensage-agent loop, polling every 5s...")

	// Enter an infinite loop to periodically collect pod information
	for {
		select {
		case <-sigCh:
			// Gracefully exit when a termination signal is received
			log.Println("Termination signal received, exiting.")
			return

		case <-ticker.C:
			// Start the pod collection in a new goroutine for asynchronous processing
			go func() {
				// Call collectOnce function to discover and log pod information
				if err := collectOnce(ctx, runtimeClient); err != nil {
					log.Printf("Error in collectOnce: %v", err)
				}
			}()
		}
	}
}

// collectOnce performs a single collection cycle by calling the Discover function
// It collects pod information and logs the details as JSON strings
func collectOnce(ctx context.Context, runtimeClient runtimeapi.RuntimeServiceClient) error {
	// Discover pods and their associated information
	podInfos, err := discovery.Discover(ctx, runtimeClient)
	if err != nil {
		return err
	}
	// Log the estimated number of pods and containers
	log.Printf("Processing %d pods, with a total of %d containers.", len(podInfos), sumContainerCount(podInfos))

	nodeInfo, err := discovery.ListNodeInfo(ctx)
	if err != nil {
		return err
	}
	// Log node info
	jsonString, _ := utils.ToJsonString(nodeInfo)
	log.Printf("NodeInfo: %v", jsonString)

	// Iterate through the discovered pod information and log it
	/*for _, podInfo := range podInfos {
		// Convert the PodInfo struct to a JSON string for logging
		jsonStr, err := utils.ToJsonString(podInfo)
		if err != nil {
			// Log an error if serializing PodInfo to JSON fails
			log.Printf("Error serializing PodInfo: %v", err)
			continue
		}
		// Log the serialized PodInfo
		log.Printf("PodInfo: %s", jsonStr)
	}*/

	return nil
}

// Helper function to sum the total number of containers across all pods
func sumContainerCount(podInfos []model.PodInfo) int {
	totalContainers := 0
	for _, podInfo := range podInfos {
		totalContainers += len(podInfo.Containers)
	}
	return totalContainers
}
