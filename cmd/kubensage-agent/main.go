// Package main implements the entry point for the kubensage-agent.
// It connects to the CRI (Container Runtime Interface), collects container metrics periodically,
// and streams them to a remote relay service over gRPC.
package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	commoncli "github.com/kubensage/go-common/cli"
	"github.com/kubensage/go-common/log"
	agentcli "github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/pkg/discovery"
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"

	"go.uber.org/zap"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const appName = "kubensage-agent"

func main() {
	logCfgLoader := commoncli.RegisterLogStdAndFileFlags(flag.CommandLine, appName)
	agentCfgLoader := agentcli.RegisterAgentFlags(flag.CommandLine)
	flag.Parse()

	logCfg := logCfgLoader()
	logger := log.SetupStdAndFileLogger(logCfg)
	defer logger.Sync()

	agentCfg := agentCfgLoader(logger)
	log.LogStartupInfo(logger, appName, logCfg, agentCfg)

	logger.Info("Connecting to relay", zap.String("relay_address", agentCfg.RelayAddress))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Warn("Shutdown signal received")
		cancel()
	}()

	logger.Info("Discovering CRI socket...")
	criSocket, err := discovery.CriSocketDiscovery()
	if err != nil {
		logger.Fatal("CRI socket discovery failed", zap.Error(err))
	}
	logger.Info("CRI socket discovered", zap.String("socket", criSocket))

	runtimeClient, criConn := utils.SetupCRIConnection(criSocket, logger)
	defer func() {
		logger.Info("Closing CRI connection")
		_ = criConn.Close()
	}()

	relayClient, relayConn := utils.SetupRelayConnection(agentCfg.RelayAddress, logger)
	defer func() {
		logger.Info("Closing relay connection")
		_ = relayConn.Close()
	}()

	bufferSize := computeBufferSize(agentCfg.MainLoopDurationSeconds, agentCfg.BufferRetention)
	buffer := utils.NewMetricsRingBuffer(bufferSize)
	logger.Info("Metrics ring buffer initialized", zap.Int("buffer_size", bufferSize))

	// SCOPED loggers
	collectorLogger := logger.Named("collector")
	senderLogger := logger.Named("sender")

	go collectionLoop(ctx, runtimeClient, buffer, agentCfg, collectorLogger)
	sendingLoop(ctx, relayClient, buffer, agentCfg, senderLogger)
}

func collectionLoop(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
	buffer *utils.RingBuffer,
	agentCfg *agentcli.AgentConfig,
	logger *zap.Logger,
) {
	logger.Info("Starting metrics collection", zap.Duration("interval", agentCfg.MainLoopDurationSeconds))

	ticker := time.NewTicker(agentCfg.MainLoopDurationSeconds)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping metrics collection")
			return

		case <-ticker.C:
			logger.Debug("Collecting metrics")
			metricsData, errs := metrics.Metrics(ctx, runtimeClient, logger, agentCfg.TopN)
			if errs != nil {
				var errStrs []string
				for _, e := range errs {
					errStrs = append(errStrs, e.Error())
				}
				logger.Error("Metric collection errors", zap.Strings("errors", errStrs))
				continue
			}
			buffer.Add(metricsData)
			logger.Debug("Metrics added to buffer", zap.Int("buffer_len", buffer.Len()))
		}
	}
}

func sendingLoop(
	ctx context.Context,
	relayClient gen.MetricsServiceClient,
	buffer *utils.RingBuffer,
	agentCfg *agentcli.AgentConfig,
	logger *zap.Logger,
) {
	logger.Info("Starting metrics sending", zap.Duration("interval", agentCfg.MainLoopDurationSeconds))

	ticker := time.NewTicker(agentCfg.MainLoopDurationSeconds)
	defer ticker.Stop()

	var stream gen.MetricsService_SendMetricsClient
	var err error

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping metrics sending")
			if stream != nil {
				if err := stream.CloseSend(); err != nil {
					logger.Warn("Error closing stream", zap.Error(err))
				} else {
					logger.Info("Stream closed")
				}
			}
			return

		case <-ticker.C:
			if stream == nil {
				logger.Debug("Opening metrics stream")
				stream, err = relayClient.SendMetrics(ctx)
				if err != nil {
					logger.Warn("Unable to open stream", zap.Error(err))
					time.Sleep(2 * time.Second)
					continue
				}
				logger.Info("Stream opened successfully")

				if err := sendAllBuffer(buffer, stream, logger); err != nil {
					logger.Error("Failed to send buffered metrics", zap.Error(err))
					stream = nil
					continue
				}
			}

			if err := popAndSend(buffer, stream, logger); err != nil {
				logger.Error("Metric send failed", zap.Error(err))
				stream = nil
			}
		}
	}
}

func sendAllBuffer(buffer *utils.RingBuffer, stream gen.MetricsService_SendMetricsClient, logger *zap.Logger) error {
	logger.Debug("Sending all buffered metrics")

	for buffer.Len() > 0 {
		if err := popAndSend(buffer, stream, logger); err != nil {
			return err
		}
	}
	return nil
}

func popAndSend(buffer *utils.RingBuffer, stream gen.MetricsService_SendMetricsClient, logger *zap.Logger) error {
	if buffer == nil {
		return errors.New("buffer is nil")
	}
	if buffer.Len() == 0 {
		logger.Debug("Buffer empty")
		return nil
	}

	pop := buffer.Pop()
	err := stream.Send(pop)
	if err != nil {
		logger.Error("Stream send failed", zap.Error(err))
		buffer.Readd(pop)
		logger.Debug("Metric re-queued", zap.Int("buffer_len", buffer.Len()))
		return err
	}

	logger.Debug("Metric sent", zap.Int("buffer_len", buffer.Len()))
	return nil
}

func computeBufferSize(loopInterval time.Duration, retentionDuration time.Duration) int {
	if loopInterval <= 0 {
		return 1
	}
	size := int(retentionDuration / loopInterval)
	if size < 1 {
		return 1
	}
	return size
}
