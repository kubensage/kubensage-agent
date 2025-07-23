package main

import (
	"context"
	"flag"
	commoncli "github.com/kubensage/go-common/cli"
	"github.com/kubensage/go-common/log"
	agentcli "github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Warn("Signal received, initiating shutdown...")
		cancel()
	}()

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
	metricsLoop(ctx, logger, runtimeClient, relayClient, stream, agentCfg.MainLoopDurationSeconds, agentCfg)
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

// metricsLoop runs the main operational loop of the kubensage agent.
//
// At each interval (defined by mainLoopDurationSeconds), it:
//   - Collects CRI-based metrics concurrently from the container runtime
//   - Converts them into protobuf format
//   - Sends them via gRPC to the relay server
//
// It handles:
//   - Stream send failures with automatic reconnection and retry
//   - Graceful shutdown via context cancellation (e.g. SIGINT/SIGTERM)
//
// This function blocks until the context is cancelled.
func metricsLoop(
	ctx context.Context,
	logger *zap.Logger,
	runtimeClient cri.RuntimeServiceClient,
	relayClient gen.MetricsServiceClient,
	stream gen.MetricsService_SendMetricsClient,
	mainLoopDurationSeconds time.Duration,
	config *agentcli.AgentConfig,
) {
	ticker := time.NewTicker(mainLoopDurationSeconds)
	defer ticker.Stop()

	bufferSize := computeBufferSize(config.MainLoopDurationSeconds, config.BufferRetention)
	buffer := metrics.NewMetricsRingBuffer(bufferSize)
	logger.Info("Ring Buffer size", zap.Int("size", bufferSize))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, cleaning up and exiting")

			if _, err := stream.CloseAndRecv(); err != nil {
				logger.Error("Failed to receive ack on shutdown", zap.Error(err))
			} else {
				logger.Info("Relay server acknowledged on shutdown")
			}
			return

		case <-ticker.C:
			// Collect CRI metrics from runtime
			collectedMetrics, errs := metrics.Metrics(ctx, runtimeClient, logger, config.TopN)
			if errs != nil {
				var errStrs []string
				for _, e := range errs {
					errStrs = append(errStrs, e.Error())
				}
				logger.Error("Failed to collect metrics", zap.Strings("errors", errStrs))
				continue
			}

			// Flush the buffer if not empty
			if buffer.Len() > 0 {
				logger.Info("Flushing buffered metrics", zap.Int("count", buffer.Len()))
				for buffer.Len() > 0 {
					m := buffer.Pop()
					if err := stream.Send(m); err != nil {
						logger.Warn("Failed to flush buffered metric", zap.Error(err))
						buffer.Add(m) // requeue the one that failed
						break
					}
					logger.Info("Flushed one buffered metric")
				}
			}

			logger.Info("Number of discovered pods", zap.Int("n_of_discovered_pods", len(collectedMetrics.PodMetrics)))
			logger.Debug("Collected metrics", zap.Any("collectedMetrics", collectedMetrics))

			// Attempt to send metrics; if stream is broken, reconnect and retry once
			if err := stream.Send(collectedMetrics); err != nil {
				logger.Warn("Stream send failed. Attempting to reconnect...", zap.Error(err))

				_ = stream.CloseSend()
				stream = openStreamWithRetry(ctx, relayClient, logger)
				if stream == nil {
					logger.Error("Stream is nil after reconnect, saving metrics to buffer")
				}

				if err2 := stream.Send(collectedMetrics); err2 != nil {
					buffer.Add(collectedMetrics)
					logger.Info("Saved current metrics to buffer", zap.Int("buffer_size", buffer.Len()))
					continue
				}
				logger.Info("Send after reconnect succeeded")
			} else {
				logger.Info("Metrics sent successfully")
			}
		}
	}
}

func computeBufferSize(loopInterval time.Duration, retentionMinutes time.Duration) int {
	size := int(retentionMinutes / loopInterval)
	if size < 1 {
		return 1
	}
	return size
}
