package main

import (
	"context"
	"github.com/kubensage/kubensage-agent/pkg/converter"
	"google.golang.org/grpc"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

// TickerDuration defines how often metrics are collected from the CRI runtime.
var TickerDuration = time.Second * 5

// main is the entry point for the kubensage-agent binary.
//
// This agent runs as a background process on Kubernetes nodes.
// It periodically collects system and container-level metrics by querying the CRI runtime via gRPC.
// Collected metrics are logged and can be forwarded by a relay to monitoring systems like Prometheus.
func main() {
	// Initialize structured file logging
	logFile := utils.SetupLogging("kubensage-agent.log")
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("error closing log file: %v", err)
		}
	}()

	// Setup signal handler for graceful shutdown (SIGINT or SIGTERM)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start periodic ticker for metric collection
	ticker := time.NewTicker(TickerDuration)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Automatically detect CRI socket (e.g., containerd or crio)
	socket, err := discovery.CriSocketDiscovery()
	if err != nil {
		log.Fatalf("Failed to discover CRI socket: %v", err)
	}

	// Establish gRPC connection to CRI runtime
	grpcCriSocketConnection := utils.AcquireGrpcConnection(socket)
	defer func(grpcConnection *grpc.ClientConn) {
		err := grpcConnection.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC connection for CRI socket: %v", err)
		}
	}(grpcCriSocketConnection)

	// Initialize CRI runtime client
	runtimeClient := runtimeapi.NewRuntimeServiceClient(grpcCriSocketConnection)

	log.Println("Starting kubensage-agent loop, polling every 5s...")

	grpcRelayConnection := utils.AcquireGrpcConnection("localhost:50051")
	defer func(grpcConnection *grpc.ClientConn) {
		err := grpcConnection.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC connection for relay: %v", err)
		}
	}(grpcRelayConnection)

	relayClient := pb.NewMetricsServiceClient(grpcRelayConnection)

	// Main loop: respond to signals or perform periodic collection
	for {
		select {
		case <-sigCh:
			log.Println("Termination signal received, exiting.")
			return

		// On each tick, collect metrics asynchronously
		case <-ticker.C:
			go func() {
				metrics, errs := discovery.GetAllMetrics(ctx, runtimeClient)

				if errs != nil {
					log.Printf("Failed to get metrics: %v", errs)
					return
				}

				converted, err := converter.ConvertToProto(metrics)
				if err != nil {
					log.Printf("Failed to convert metrics: %v", err)
					return
				}

				ack, err := relayClient.SendMetrics(ctx, converted)
				if err != nil {
					log.Printf("Error sending metrics to relay server: %v", err)
					return
				}
				log.Printf("Relay server acknowledged: %s", ack.Message)

				// Serialize metrics as JSON for debugging/logging
				/*jsonStr, _ := utils.ToJsonString(metrics)
				log.Println(jsonStr)*/
			}()
		}
	}
}
