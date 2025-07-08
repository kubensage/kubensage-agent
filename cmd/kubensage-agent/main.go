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
	"runtime"
	"syscall"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

type config struct {
	relayAddress            string
	mainLoopDurationSeconds time.Duration

	logLevel      string
	logFile       string
	logMaxSize    int
	logMaxBackups int
	logMaxAge     int
	logCompress   bool
}

func main() {
	flags := parseFlags()
	logger := setupLogger(flags)

	// Log basic runtime info on agent startup
	logger.Info("kubensage-agent started", zap.String("version", runtime.Version()),
		zap.Time("start_time", time.Now()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensures all context-aware operations can exit cleanly

	sigCh := handleSignal()

	logger.Info("Discovering CRI socket")
	criSocket, err := discovery.CriSocketDiscovery()
	if err != nil {
		logger.Fatal("Failed to discover CRI socket", zap.Error(err))
	}
	logger.Info("Discovered CRI socket", zap.String("socket", criSocket))

	// Connect to CRI and defer cleanup of connection
	runtimeClient, criConn := setupCRIConnection(criSocket, logger)
	defer func(criConn *grpc.ClientConn) {
		err := criConn.Close()
		if err != nil {
			logger.Warn("Failed to close CRI connection", zap.Error(err))
		}
	}(criConn)

	// Connect to relay and defer cleanup of connection
	relayClient, relayConn := setupRelayConnection(flags.relayAddress, logger)
	defer func(relayConn *grpc.ClientConn) {
		err := relayConn.Close()
		if err != nil {
			logger.Warn("Failed to close relay connection", zap.Error(err))
		}
	}(relayConn)

	logger.Info("Opening initial stream channel")
	stream := openStreamWithRetry(ctx, relayClient, logger)

	// Start the core metric collection loop
	metricsLoop(ctx, logger, runtimeClient, relayClient, stream, sigCh, flags.mainLoopDurationSeconds)
}

func setupLogger(flags config) *zap.Logger {
	logger, err := utils.NewLogger(&flags.logLevel, &flags.logFile, &flags.logMaxSize, &flags.logMaxBackups,
		&flags.logMaxAge, &flags.logCompress,
	)

	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	return logger
}

func parseFlags() config {
	relayAddress := flag.String("relay-address", "",
		"The address of the relay grpc server, (Required: yes, Default: N/A)")

	mainLoopDurationSecondsFlag := flag.Int("main-loop-duration-seconds", 5,
		"The duration of the main loop (Required: No, Default: 5s)")

	logLevel := flag.String("log-level", "info",
		"Set log level, (Required: No, Default: info)")

	logFile := flag.String("log-file", "/var/log/kubensage/kubensage-agent.log",
		"Path to log file, (Required: No, Default: /var/log/kubensage-agent.log)")

	logMaxSize := flag.Int("log-max-size", 10,
		"Maximum log size (MB), (Required: No, Default: 10)")

	logMaxBackups := flag.Int("log-max-backups", 5,
		"Max backup files, (Required: No, Default: 5)")

	logMaxAge := flag.Int("log-max-age", 30,
		"Max age in days to retain old log files, (Required: No, Default: 30)")

	logCompress := flag.Bool("log-compress", true,
		"Compress old log files, (Required: No, Default: true)")

	flag.Parse()

	mainLoopDuration := time.Duration(*mainLoopDurationSecondsFlag) * time.Second

	if *relayAddress == "" {
		log.Fatal("Missing required flag: --relay-address")
	}

	return config{
		relayAddress:            *relayAddress,
		mainLoopDurationSeconds: mainLoopDuration,
		logLevel:                *logLevel,
		logFile:                 *logFile,
		logMaxSize:              *logMaxSize,
		logMaxBackups:           *logMaxBackups,
		logMaxAge:               *logMaxAge,
		logCompress:             *logCompress,
	}
}

func setupCRIConnection(
	socket string,
	logger *zap.Logger,
) (client runtimeapi.RuntimeServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to CRI socket", zap.String("socket", socket))
	conn := utils.AcquireGrpcConnection(socket, logger)
	logger.Info("Connected to CRI socket")
	return runtimeapi.NewRuntimeServiceClient(conn), conn
}

func setupRelayConnection(
	addr string,
	logger *zap.Logger,
) (client pb.MetricsServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := utils.AcquireGrpcConnection(addr, logger)
	logger.Info("Connected to relay GRPC server")
	return pb.NewMetricsServiceClient(conn), conn
}

// openStreamWithRetry attempts to establish a streaming connection with the relay server using a retry loop.
// It implements an exponential backoff strategy:
// - Starts with a 1-second wait between retries.
// - On each failure, the wait time doubles (1s → 2s → 4s → 8s, etc.).
// - The backoff duration is capped at 30 seconds to prevent excessively long delays.
// - If the context is cancelled (e.g., due to shut down), the function logs and returns nil.
// This mechanism helps reduce pressure on the relay server during outages or instability.
func openStreamWithRetry(
	ctx context.Context,
	client pb.MetricsServiceClient,
	logger *zap.Logger,
) pb.MetricsService_SendMetricsClient {

	backoff := time.Second

	for {
		logger.Info("Opening stream to relay server...")
		stream, err := client.SendMetrics(ctx)
		if err == nil {
			logger.Info("Stream opened successfully")
			return stream
		}
		logger.Error("Failed to open stream, retrying...", zap.Error(err))

		select {
		case <-ctx.Done():
			// If the context is cancelled externally (e.g. SIGTERM or shutdown), stop retrying
			logger.Warn("Context cancelled during stream reconnect")
			return nil
		case <-time.After(backoff):
			// Wait for the current backoff duration before retrying
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func handleSignal() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}

// metricsLoop runs the main operational loop of the agent.
// At each interval:
// - It collects CRI-based metrics concurrently
// - Converts them into protobuf format
// - Sends them via GRPC to the relay server
// It also handles reconnection on stream failure and responds to shut down signals.

func metricsLoop(
	ctx context.Context,
	logger *zap.Logger,
	runtimeClient runtimeapi.RuntimeServiceClient,
	relayClient pb.MetricsServiceClient,
	stream pb.MetricsService_SendMetricsClient,
	sigCh <-chan os.Signal,
	mainLoopDurationSeconds time.Duration,
) {
	ticker := time.NewTicker(mainLoopDurationSeconds)
	defer ticker.Stop() // Ensure ticker doesn't leak if function exits

	for {
		select {
		case <-sigCh:
			// Signal received: close the stream and exit
			ack, err := stream.CloseAndRecv()
			if err != nil {
				logger.Error("Failed to receive ack", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged", zap.String("relay_response", ack.Message))
			}
			logger.Warn("Termination signal received, exiting")
			return

		case <-ctx.Done():
			// Context cancelled externally: close the stream and exit
			ack, err := stream.CloseAndRecv()
			if err != nil {
				logger.Error("Failed to receive ack on context cancel", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged on cancel", zap.String("relay_response", ack.Message))
			}
			logger.Info("Context cancelled, exiting")
			return

		case <-ticker.C:
			// Triggered by ticker: collect and send metrics

			metrics, errs := discovery.GetAllMetrics(ctx, runtimeClient, logger)
			if errs != nil {
				var errStrs []string
				for _, e := range errs {
					errStrs = append(errStrs, e.Error())
				}
				logger.Error("Failed to get metrics", zap.Strings("errors", errStrs))
				continue
			}

			converted, err := converter.ConvertToProto(metrics)
			if err != nil {
				logger.Error("Failed to convert metrics", zap.Error(err))
				continue
			}

			// Attempt to send metrics; on failure, reconnect and retry once
			if err := stream.Send(converted); err != nil {
				logger.Warn("Stream send failed. Attempting to reconnect...", zap.Error(err))

				_ = stream.CloseSend() // Ensure we explicitly close the failed stream

				stream = openStreamWithRetry(ctx, relayClient, logger)
				logger.Info("Reconnected to stream successfully")

				err2 := stream.Send(converted)
				if err2 != nil {
					logger.Error("Send after reconnect failed", zap.Error(err2))
					continue
				}

				logger.Info("Send after reconnect succeeded")
			} else {
				logger.Info("Metrics sent successfully", zap.Int("n_of_discovered_pods", len(converted.PodMetrics)))
			}
		}
	}
}
