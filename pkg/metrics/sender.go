package metrics

import (
	"context"
	"errors"
	agentcli "github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"time"
)

func SendingLoop(
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
