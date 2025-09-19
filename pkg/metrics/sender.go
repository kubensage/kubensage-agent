package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/kubensage/go-common/datastructure"
	"github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
)

// SendOnce attempts to send one batch of metrics from the ring buffer over the provided gRPC stream.
//
// If the stream is nil, it opens a new gRPC stream using the provided MetricsServiceClient.
// Before sending the next metric, it first flushes the entire buffer using sendAllBuffer().
// If any error occurs while opening the stream or sending data, it logs the error and returns it,
// ensuring the stream is marked as unusable (by setting it to nil) for retry on the next call.
//
// Parameters:
//   - ctx context.Context:
//     Context used to manage deadlines and cancellation during the stream lifecycle.
//   - relayClient gen.MetricsServiceClient:
//     gRPC client used to create the metrics stream.
//   - stream gen.MetricsService_SendMetricsClient:
//     Current active stream used for sending metrics. If nil, a new one will be created.
//   - buffer *datastructure.RingBuffer[*gen.Metrics]:
//     Ring buffer containing collected metrics that need to be sent.
//   - agentCfg *cli.AgentConfig:
//     Agent configuration used for accessing loop timing (used in retries).
//   - logger *zap.Logger:
//     Structured logger for debug and error output.
//
// Returns:
//   - error:
//     An error is returned if opening the stream or sending metrics fails. The caller
//     is responsible for handling the failed stream (usually by retrying on the next loop).
func SendOnce(
	ctx context.Context,
	relayClient gen.MetricsServiceClient,
	stream gen.MetricsService_SendMetricsClient,
	buffer *datastructure.RingBuffer[*gen.Metrics],
	agentCfg *cli.AgentConfig,
	logger *zap.Logger,
) error {
	start := time.Now()

	var err error
	if stream == nil {
		logger.Info("opening metrics stream")
		stream, err = relayClient.SendMetrics(ctx)
		if err != nil {
			logger.Warn("unable to open stream", zap.Error(err))
			time.Sleep(agentCfg.MainLoopDurationSeconds)
			return err
		}
		logger.Info("stream opened successfully")
	}

	// Prima di inviare il prossimo, flush dell’intero buffer
	flushStart := time.Now()
	flushed, err := sendAllBuffer(buffer, stream, logger)
	if err != nil {
		logger.Error("failed to send buffered metrics",
			zap.Int("flushed_before_error", flushed),
			zap.Duration("flush_duration", time.Since(flushStart)),
			zap.Error(err),
		)

		stream = nil
		return err
	}

	sendStart := time.Now()
	if err := popAndSend(buffer, stream, logger); err != nil {
		logger.Error("metric send failed",
			zap.Duration("send_duration", time.Since(sendStart)),
			zap.Error(err),
		)
		stream = nil
		return err
	}

	logger.Info("metrics send cycle completed",
		zap.Int("flushed_count", flushed),
		zap.Duration("flush_duration", time.Since(flushStart)),
		zap.Duration("cycle_total_duration", time.Since(start)),
	)

	return nil
}

// sendAllBuffer attempts to flush all pending metrics from the buffer through the gRPC stream.
//
// It repeatedly calls popAndSend() until the buffer is empty or an error occurs.
// If sending a metric fails, the function stops immediately and returns the error.
// This ensures reliability by avoiding partial or failed batches.
//
// Parameters:
//   - buffer *datastructure.RingBuffer[*gen.Metrics]:
//     A ring buffer containing queued metrics to be sent.
//   - stream gen.MetricsService_SendMetricsClient:
//     gRPC client stream used for sending metrics to the relay service.
//   - logger *zap.Logger:
//     Logger for debug and error messages.
//
// Returns:
//   - error:
//     Returns the first error encountered while sending, or nil if all metrics were flushed successfully.
func sendAllBuffer(
	buffer *datastructure.RingBuffer[*gen.Metrics],
	stream gen.MetricsService_SendMetricsClient,
	logger *zap.Logger,
) (int, error) {
	if buffer == nil {
		return 0, errors.New("buffer is nil")
	}

	if buffer.Len() == 0 {
		logger.Debug("buffer empty (nothing to flush)")
		return 0, nil
	}

	logger.Info("flushing buffered metrics", zap.Int("buffer_len_start", buffer.Len()))
	start := time.Now()

	count := 0
	for buffer.Len() > 0 {
		if err := popAndSend(buffer, stream, logger); err != nil {
			return count, err
		}
		count++
	}

	logger.Info("buffer flushed",
		zap.Int("flushed_count", count),
		zap.Duration("flush_duration", time.Since(start)),
	)
	return count, nil
}

// popAndSend attempts to send a single Metrics message from the ring buffer over a gRPC stream.
//
// It pops the oldest metric from the buffer and calls stream.Send() to transmit it.
// If the buffer is empty, the function exits silently. If sending fails, the metric is re-added
// to the buffer for retry, and the error is returned.
//
// This function ensures metrics are not lost in case of transient gRPC failures.
//
// Parameters:
//   - buffer *datastructure.RingBuffer[*gen.Metrics]:
//     A ring buffer holding pending metrics to be sent.
//   - stream gen.MetricsService_SendMetricsClient:
//     gRPC client stream used to send metrics to the relay service.
//   - logger *zap.Logger:
//     Logger for debugging and error logging.
//
// Returns:
//   - error:
//     Returns an error if the buffer is nil or sending over the stream fails.
//     Returns nil if the buffer is empty or the metric is sent successfully.
func popAndSend(
	buffer *datastructure.RingBuffer[*gen.Metrics],
	stream gen.MetricsService_SendMetricsClient,
	logger *zap.Logger,
) error {
	if buffer == nil {
		return errors.New("buffer is nil")
	}
	if buffer.Len() == 0 {
		logger.Debug("buffer empty")
		return nil
	}

	_, pop, ok := buffer.Pop()
	if !ok {
		logger.Debug("buffer empty after check (race condition?)")
		return nil
	}

	if err := stream.Send(pop); err != nil {
		logger.Error("stream send failed", zap.Error(err))
		buffer.Readd(pop)
		logger.Debug("metric re-queued", zap.Int("buffer_len", buffer.Len()))
		return err
	}

	logger.Debug("metric sent", zap.Int("buffer_len", buffer.Len()))
	return nil
}
