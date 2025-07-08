package main

import (
	"context"
	"flag"
	"github.com/kubensage/kubensage-agent/pkg/converter"
	"go.uber.org/zap"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := handleSignal()

	logger.Info("Discovering CRI socket")
	criSocket, err := discovery.CriSocketDiscovery()
	if err != nil {
		logger.Fatal("Failed to discover CRI socket", zap.Error(err))
	}
	logger.Info("Discovered CRI socket", zap.String("socket", criSocket))

	runtimeClient := setupCRIConnection(criSocket, logger)
	relayClient := setupRelayConnection(flags.relayAddress, logger)

	logger.Info("Opening initial stream channel")
	stream := openStreamWithRetry(ctx, relayClient, logger)

	startMetricsLoop(ctx, logger, runtimeClient, relayClient, stream, sigCh, flags.mainLoopDurationSeconds)
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
	relayAddress := flag.String("relay-address", "", "The address of the relay grpc server, (Required: yes, Default: N/A)")
	mainLoopDurationSecondsFlag := flag.Int("main-loop-duration-seconds", 5, "The duration of the main loop (Required: No, Default: 5s)")
	mainLoopDuration := time.Duration(*mainLoopDurationSecondsFlag) * time.Second

	logLevel := flag.String("log-level", "info", "Set log level, (Required: No, Default: info)")
	logFile := flag.String("log-file", "/var/log/kubensage/kubensage-agent.log", "Path to log file, (Required: No, Default: /var/log/kubensage-agent.log)")
	logMaxSize := flag.Int("log-max-size", 10, "Maximum log size (MB), (Required: No, Default: 10)")
	logMaxBackups := flag.Int("log-max-backups", 5, "Max backup files, (Required: No, Default: 5)")
	logMaxAge := flag.Int("log-max-age", 30, "Max age in days to retain old log files, (Required: No, Default: 30)")
	logCompress := flag.Bool("log-compress", true, "Compress old log files, (Required: No, Default: true)")

	flag.Parse()

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

func setupCRIConnection(socket string, logger *zap.Logger) runtimeapi.RuntimeServiceClient {
	logger.Info("Connecting to CRI socket", zap.String("socket", socket))
	conn := utils.AcquireGrpcConnection(socket, *logger)
	logger.Info("Connected to CRI socket")
	return runtimeapi.NewRuntimeServiceClient(conn)
}

func setupRelayConnection(addr string, logger *zap.Logger) pb.MetricsServiceClient {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := utils.AcquireGrpcConnection(addr, *logger)
	logger.Info("Connected to relay GRPC server")
	return pb.NewMetricsServiceClient(conn)
}

func openStreamWithRetry(ctx context.Context, client pb.MetricsServiceClient, logger *zap.Logger) pb.MetricsService_SendMetricsClient {
	for {
		logger.Info("Opening stream to relay server...")
		stream, err := client.SendMetrics(ctx)
		if err == nil {
			logger.Info("Stream opened successfully")
			return stream
		}
		logger.Error("Failed to open stream, retrying...", zap.Error(err))
		time.Sleep(2 * time.Second)
	}
}

func handleSignal() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}

func startMetricsLoop(
	ctx context.Context,
	logger *zap.Logger,
	runtimeClient runtimeapi.RuntimeServiceClient,
	relayClient pb.MetricsServiceClient,
	stream pb.MetricsService_SendMetricsClient,
	sigCh <-chan os.Signal,
	mainLoopDurationSeconds time.Duration,
) {
	ticker := time.NewTicker(mainLoopDurationSeconds)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			ack, err := stream.CloseAndRecv()
			if err != nil {
				logger.Error("Failed to receive ack", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged", zap.String("relay_response", ack.Message))
			}
			logger.Warn("Termination signal received, exiting")
			return

		case <-ctx.Done():
			logger.Info("Context cancelled, exiting")
			return

		case <-ticker.C:
			metrics, errs := discovery.GetAllMetrics(ctx, runtimeClient, *logger)
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

			if err := stream.Send(converted); err != nil {
				logger.Warn("Stream send failed. Attempting to reconnect...", zap.Error(err))

				stream = openStreamWithRetry(ctx, relayClient, logger)
				logger.Info("Reconnected to stream successfully")

				if err := stream.Send(converted); err != nil {
					logger.Error("Send after reconnect failed", zap.Error(err))
					continue
				}

				logger.Info("Send after reconnect succeeded")
			} else {
				logger.Info("Metrics sent successfully", zap.Int("n_of_discovered_pods", len(converted.PodMetrics)))
			}

		}
	}
}
