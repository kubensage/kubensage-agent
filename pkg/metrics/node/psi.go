package node

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// buildPsiMetrics reads Pressure Stall Information (PSI) metrics from a given /proc file.
//
// PSI metrics provide visibility into resource pressure for CPU, memory, and I/O.
// The file is expected to follow the format of /proc/pressure/{cpu|memory|io}.
// This function parses `some` and `full` stall entries, extracting avg10, avg60,
// avg300, and total values into a gen.PsiMetrics protobuf message.
//
// If the file cannot be read or parsed, a zero-valued PsiMetrics struct is returned.
//
// Parameters:
//   - path: Absolute path to the PSI file (e.g., "/proc/pressure/cpu")
//   - logger: Logger used for debug and error tracing
//
// Returns:
//   - *gen.PsiMetrics containing the parsed stall data
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func buildPsiMetrics(
	path string,
	logger *zap.Logger,
) (*gen.PsiMetrics, time.Duration) {
	start := time.Now()

	file, err := os.Open(path)
	if err != nil {
		logger.Error("failed to open metrics file", zap.String("path", path), zap.Error(err))
		return &gen.PsiMetrics{}, time.Since(start)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			logger.Error("error closing PSI file", zap.String("path", path), zap.Error(err))
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var metrics gen.PsiMetrics

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 5 {
			continue
		}

		entry := &gen.PsiData{}

		for _, part := range parts[1:] {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				continue
			}

			switch kv[0] {
			case "avg10":
				if val, err := strconv.ParseFloat(kv[1], 64); err == nil {
					entry.Avg10 = wrapperspb.Double(val)
				}
			case "avg60":
				if val, err := strconv.ParseFloat(kv[1], 64); err == nil {
					entry.Avg60 = wrapperspb.Double(val)
				}
			case "avg300":
				if val, err := strconv.ParseFloat(kv[1], 64); err == nil {
					entry.Avg300 = wrapperspb.Double(val)
				}
			case "total":
				if val, err := strconv.ParseUint(kv[1], 10, 64); err == nil {
					entry.Total = wrapperspb.UInt64(val)
				}
			}
		}

		switch parts[0] {
		case "some":
			metrics.Some = entry
		case "full":
			metrics.Full = entry
		}
	}

	return &metrics, time.Since(start)
}
