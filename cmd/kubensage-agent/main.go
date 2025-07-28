// Package main implements the entry point for the kubensage-agent.
// It connects to the CRI (Container Runtime Interface), collects container metrics periodically,
// and streams them to a remote relay service over gRPC.
package main

import (
	"context"
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
	"go.uber.org/zap"
)

const appName = "kubensage-agent"

func main() {
	logCfgLoader := commoncli.RegisterLogStdAndFileFlags(flag.CommandLine, appName)
	agentCfgLoader := agentcli.RegisterAgentFlags(flag.CommandLine)
	flag.Parse()

	logCfg := logCfgLoader()
	logger := log.SetupStdAndFileLogger(logCfg)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Fatal("failed to sync logger", zap.Error(err))
		}
	}(logger)

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
	logger.Info("metrics ring buffer initialized", zap.Int("buffer_size", bufferSize))

	// SCOPED loggers
	collectorLogger := logger.Named("collector")
	senderLogger := logger.Named("sender")

	go func() {
		ticker := time.NewTicker(agentCfg.MainLoopDurationSeconds)
		defer ticker.Stop()

		stream, err := relayClient.SendMetrics(ctx)
		if err != nil {
			logger.Error("failed to send metrics", zap.Error(err))
		} else {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					errors := metrics.CollectOnce(ctx, runtimeClient, buffer, agentCfg, collectorLogger)

					if errors != nil {
						logger.Error("Errors while collecting metrics", zap.Any("errors", errors))
						return
					}

					err := metrics.SendOnce(ctx, relayClient, stream, buffer, senderLogger)
					if err != nil {
						logger.Error("Error while sending metrics", zap.Error(err))
						return
					}
				}
			}
		}

	}()
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
