package main

import (
	"context"
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
)

func main() {
	flags := utils.ParseFlags()
	logger := utils.SetupLogger(flags)

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
	runtimeClient, criConn := utils.SetupCRIConnection(criSocket, logger)
	defer func(criConn *grpc.ClientConn) {
		err := criConn.Close()
		if err != nil {
			logger.Warn("Failed to close CRI connection", zap.Error(err))
		}
	}(criConn)

	// Connect to relay and defer cleanup of connection
	relayClient, relayConn := utils.SetupRelayConnection(flags.RelayAddress, logger)
	defer func(relayConn *grpc.ClientConn) {
		err := relayConn.Close()
		if err != nil {
			logger.Warn("Failed to close relay connection", zap.Error(err))
		}
	}(relayConn)

	logger.Info("Opening initial stream channel")
	stream := openStreamWithRetry(ctx, relayClient, logger)

	// Start the core metric collection loop
	metricsLoop(ctx, logger, runtimeClient, relayClient, stream, sigCh, flags.MainLoopDurationSeconds)
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
	client proto.MetricsServiceClient,
	logger *zap.Logger,
) proto.MetricsService_SendMetricsClient {

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

// metricsLoop runs the main operational loop of the agent.
// At each interval:
// - It collects CRI-based metrics concurrently
// - Converts them into protobuf format
// - Sends them via GRPC to the relay server
// It also handles reconnection on stream failure and responds to shut down signals.
func metricsLoop(
	ctx context.Context,
	logger *zap.Logger,
	runtimeClient cri.RuntimeServiceClient,
	relayClient proto.MetricsServiceClient,
	stream proto.MetricsService_SendMetricsClient,
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

			metrics, errs := metrics.GetMetrics(ctx, runtimeClient, logger)
			if errs != nil {
				var errStrs []string
				for _, e := range errs {
					errStrs = append(errStrs, e.Error())
				}
				logger.Error("Failed to get metrics", zap.Strings("errors", errStrs))
				continue
			}

			logger.Debug("Metrics", zap.Any("metrics", metrics))

			// Attempt to send metrics; on failure, reconnect and retry once
			if err := stream.Send(metrics); err != nil {
				logger.Warn("Stream send failed. Attempting to reconnect...", zap.Error(err))

				_ = stream.CloseSend() // Ensure we explicitly close the failed stream

				stream = openStreamWithRetry(ctx, relayClient, logger)
				logger.Info("Reconnected to stream successfully")

				err2 := stream.Send(metrics)
				if err2 != nil {
					logger.Error("Send after reconnect failed", zap.Error(err2))
					continue
				}

				logger.Info("Send after reconnect succeeded")
			} else {
				logger.Info("Metrics sent successfully", zap.Int("n_of_discovered_pods", len(metrics.PodMetrics)))
			}
		}
	}
}

func handleSignal() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}
