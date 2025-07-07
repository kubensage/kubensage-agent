package main

import (
	"context"
	"flag"
	"github.com/kubensage/kubensage-agent/pkg/converter"
	"go.uber.org/zap"
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
var RelayGrpcServerAddress = "localhost:50051"

// main is the entry point for the kubensage-agent binary.
//
// This agent runs as a background process on Kubernetes nodes.
// It periodically collects system and container-level metrics by querying the CRI runtime via gRPC.
// Collected metrics are logged and can be forwarded by a relay to monitoring systems like Prometheus.
func main() {
	// === Logging flags ===
	logLevel := flag.String("log-level", "info", "Set log level: debug, info, warn, error")
	logFile := flag.String("log-file", "/var/log/kubensage/kubensage-agent.log", "Path to log file")
	logMaxSize := flag.Int("log-max-size", 10, "Maximum size in MB before rotation")
	logMaxBackups := flag.Int("log-max-backups", 5, "Maximum number of old log files to retain")
	logMaxAge := flag.Int("log-max-age", 30, "Maximum number of days to retain old log files")
	logCompress := flag.Bool("log-compress", true, "Whether to compress rotated log files")

	flag.Parse()

	// Setup logging
	logger, err := utils.NewLogger(logLevel, logFile, logMaxSize, logMaxBackups, logMaxAge, logCompress)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Fatalf("Failed to sync logger: %v", err)
		}
	}(logger)

	logger.Info("Starting Kubensage Agent", zap.String("log_level", logger.Level().String()))
	logger.Info("Log configuration",
		zap.String("file", *logFile),
		zap.Int("max_size_mb", *logMaxSize),
		zap.Int("max_backups", *logMaxBackups),
		zap.Int("max_age_days", *logMaxAge),
		zap.Bool("compress", *logCompress),
	)

	// Setup signal handler for graceful shutdown (SIGINT or SIGTERM)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start periodic ticker for metric collection
	ticker := time.NewTicker(TickerDuration)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Automatically detect CRI socket (e.g., containerd or crio)
	logger.Info("Discovering CRI socket")
	socket, err := discovery.CriSocketDiscovery()
	if err != nil {
		logger.Fatal("Failed to discover CRI socket", zap.Error(err))
	}
	logger.Info("Discovered CRI socket", zap.String("socket", socket))

	// Establish gRPC connection to CRI runtime
	logger.Info("Establishing connection to CRI socket", zap.String("socket", socket))
	grpcCriSocketConnection := utils.AcquireGrpcConnection(socket, *logger)
	defer func(grpcConnection *grpc.ClientConn) {
		err := grpcConnection.Close()
		if err != nil {
			logger.Fatal("Failed to close gRPC connection for CRI socket", zap.Error(err))
		}
	}(grpcCriSocketConnection)
	logger.Info("Connected to CRI socket", zap.String("socket", socket))

	// Initialize CRI runtime client
	runtimeClient := runtimeapi.NewRuntimeServiceClient(grpcCriSocketConnection)

	logger.Info("Establishing connection to relay GRPC server", zap.String("socket", RelayGrpcServerAddress))
	grpcRelayConnection := utils.AcquireGrpcConnection(RelayGrpcServerAddress, *logger)
	defer func(grpcConnection *grpc.ClientConn) {
		err := grpcConnection.Close()
		if err != nil {
			logger.Fatal("Failed to close gRPC connection for relay", zap.Error(err))
		}
	}(grpcRelayConnection)
	logger.Info("Connected to relay GRPC server", zap.String("socket", RelayGrpcServerAddress))

	relayClient := pb.NewMetricsServiceClient(grpcRelayConnection)

	// Main loop: respond to signals or perform periodic collection
	logger.Info("Starting collection loop", zap.Int("interval_seconds", int(TickerDuration.Seconds())))
	for {
		select {
		case <-sigCh:
			logger.Warn("Stopping Kubensage Agent, termination signal received")
			return

		// On each tick, collect metrics asynchronously
		case <-ticker.C:
			go func() {
				metrics, errs := discovery.GetAllMetrics(ctx, runtimeClient, *logger)

				if errs != nil {
					var errStrs []string
					for _, e := range errs {
						errStrs = append(errStrs, e.Error())
					}
					logger.Error("Failed to get metrics", zap.Strings("errors", errStrs))
					return
				}

				logger.Debug("Got metrics", zap.Any("metrics", metrics))

				converted, err := converter.ConvertToProto(metrics)
				if err != nil {
					logger.Error("Failed to convert metrics", zap.Error(err))
					return
				}

				ack, err := relayClient.SendMetrics(ctx, converted)
				if err != nil {
					logger.Error("Error sending metrics to relay server", zap.Error(err))
					return
				}
				logger.Info("Relay server acknowledged", zap.String("relay_response", ack.Message))
			}()
		}
	}
}
