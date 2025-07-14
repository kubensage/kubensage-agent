package main

import (
	"context"
	"flag"
	commoncli "gitlab.com/kubensage/go-common/cli"
	"gitlab.com/kubensage/go-common/log"
	agentcli "gitlab.com/kubensage/kubensage-agent/pkg/cli"
	"gitlab.com/kubensage/kubensage-agent/pkg/discovery"
	"gitlab.com/kubensage/kubensage-agent/pkg/metrics"
	"gitlab.com/kubensage/kubensage-agent/pkg/utils"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const appName = "kubensage-agent"

func main() {
	logCfgFn := commoncli.RegisterLogStdAndFileFlags(flag.CommandLine, appName)
	agentCfgFn := agentcli.RegisterAgentFlags(flag.CommandLine)

	flag.Parse()

	logCfg := logCfgFn()
	logger := log.SetupStdAndFileLogger(logCfg)
	agentCfg := agentCfgFn(logger)

	log.LogStartupInfo(logger, appName, logCfg, agentCfg)

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
	relayClient, relayConn := utils.SetupRelayConnection(agentCfg.RelayAddress, logger)
	defer func(relayConn *grpc.ClientConn) {
		err := relayConn.Close()
		if err != nil {
			logger.Warn("Failed to close relay connection", zap.Error(err))
		}
	}(relayConn)

	logger.Info("Opening initial stream channel")
	stream := openStreamWithRetry(ctx, relayClient, logger)

	// Start the core metric collection loop
	metricsLoop(ctx, logger, runtimeClient, relayClient, stream, sigCh, agentCfg.MainLoopDurationSeconds)
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
	client gen.MetricsServiceClient,
	logger *zap.Logger,
) gen.MetricsService_SendMetricsClient {

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
	relayClient gen.MetricsServiceClient,
	stream gen.MetricsService_SendMetricsClient,
	sigCh <-chan os.Signal,
	mainLoopDurationSeconds time.Duration,
) {
	ticker := time.NewTicker(mainLoopDurationSeconds)
	defer ticker.Stop() // Ensure ticker doesn't leak if function exits

	for {
		select {
		case <-sigCh:
			// Signal received: close the stream and exit
			_, err := stream.CloseAndRecv()
			if err != nil {
				logger.Error("Failed to receive ack", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged")
			}
			logger.Warn("Termination signal received, exiting")
			return

		case <-ctx.Done():
			// Context cancelled externally: close the stream and exit
			_, err := stream.CloseAndRecv()
			if err != nil {
				logger.Error("Failed to receive ack on context cancel", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged on cancel")
			}
			logger.Info("Context cancelled, exiting")
			return

		case <-ticker.C:
			// Triggered by ticker: collect and send collectedMetrics

			collectedMetrics, errs := metrics.Metrics(ctx, runtimeClient, logger)
			if errs != nil {
				var errStrs []string
				for _, e := range errs {
					errStrs = append(errStrs, e.Error())
				}
				logger.Error("Failed to get collectedMetrics", zap.Strings("errors", errStrs))
				continue
			}

			logger.Info("Number of discovered pods", zap.Int("n_of_discovered_pods", len(collectedMetrics.PodMetrics)))
			logger.Debug("Metrics", zap.Any("collectedMetrics", collectedMetrics))

			// Attempt to send collectedMetrics; on failure, reconnect and retry once
			if err := stream.Send(collectedMetrics); err != nil {
				logger.Warn("Stream send failed. Attempting to reconnect...", zap.Error(err))

				_ = stream.CloseSend() // Ensure we explicitly close the failed stream

				stream = openStreamWithRetry(ctx, relayClient, logger)
				logger.Info("Reconnected to stream successfully")

				err2 := stream.Send(collectedMetrics)
				if err2 != nil {
					logger.Error("Send after reconnect failed", zap.Error(err2))
					continue
				}

				logger.Info("Send after reconnect succeeded")
			} else {
				logger.Info("Metrics sent successfully")
			}
		}
	}
}

func handleSignal() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}
