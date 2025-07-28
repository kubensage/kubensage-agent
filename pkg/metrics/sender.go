package metrics

import (
	"context"
	"errors"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"time"
)

// SendOnce starts a continuous loop that establishes a gRPC stream to the relay server
// and sends collected metrics from the ring buffer.
//
// Parameters:
// - ctx: Context used to control the lifecycle of the loop (cancellation, shutdown).
// - relayClient: gRPC client used to send metrics to the remote backend.
// - buffer: Ring buffer holding collected metrics waiting to be sent.
// - agentCfg: Agent configuration that defines the send interval.
// - logger: Zap logger for structured logging.
//
// Behavior:
// - On each tick (based on the configured interval), it ensures a stream is open.
// - If the stream is nil or broken, it tries to reopen it.
// - If reopening is successful, it first flushes all previously buffered metrics.
// - Then it attempts to pop and send the latest metric.
// - If a send fails, the stream is closed and retried on the next tick.
// - The loop exits gracefully when the context is cancelled.
func SendOnce(
	ctx context.Context,
	relayClient gen.MetricsServiceClient,
	stream gen.MetricsService_SendMetricsClient,
	buffer *utils.RingBuffer,
	logger *zap.Logger,
) error {
	var err error

	if stream == nil {
		logger.Debug("Opening metrics stream")
		stream, err = relayClient.SendMetrics(ctx)
		if err != nil {
			logger.Warn("Unable to open stream", zap.Error(err))
			time.Sleep(2 * time.Second)
			return err
		}
		logger.Info("Stream opened successfully")

		if err := sendAllBuffer(buffer, stream, logger); err != nil {
			logger.Error("Failed to send buffered metrics", zap.Error(err))
			stream = nil
			return err
		}
	}

	if err := popAndSend(buffer, stream, logger); err != nil {
		logger.Error("Metric send failed", zap.Error(err))
		stream = nil
	}

	return nil
}

// sendAllBuffer attempts to flush the entire contents of the ring buffer over the provided gRPC stream.
//
// Parameters:
// - buffer: The ring buffer containing metrics to be sent.
// - stream: The active gRPC client stream.
// - logger: Logger used to record operation status.
//
// Returns:
// - An error if sending a metric fails (in which case remaining items stay in the buffer).
// - Nil if all metrics are sent successfully.
func sendAllBuffer(buffer *utils.RingBuffer, stream gen.MetricsService_SendMetricsClient, logger *zap.Logger) error {
	logger.Debug("Sending all buffered metrics")

	for buffer.Len() > 0 {
		if err := popAndSend(buffer, stream, logger); err != nil {
			return err
		}
	}
	return nil
}

// popAndSend pops the oldest metric from the buffer and sends it over the gRPC stream.
// If sending fails, the metric is reinserted into the buffer for retry.
//
// Parameters:
// - buffer: The ring buffer holding metrics.
// - stream: The gRPC client stream to send metrics through.
// - logger: Logger used for debug and error output.
//
// Returns:
// - An error if the send operation fails.
// - Nil if the metric is successfully sent or if the buffer is empty.
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
