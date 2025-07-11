package metrics

import (
	"bufio"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
)

// SafePsiMetrics reads and parses pressure stall information (PSI) from the given /proc/pressure/<resource> file.
// It safely extracts "some" and "full" pressure lines and their average/total values.
// If the file cannot be opened or parsed, it logs the error and returns an empty PsiMetrics struct.
func SafePsiMetrics(path string, logger *zap.Logger) *proto.PsiMetrics {
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open metrics file", zap.String("path", path), zap.Error(err))
		// Return zero-value metrics but continue
		return &proto.PsiMetrics{}
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			logger.Error("Error closing PSI file", zap.String("path", path), zap.Error(err))
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var metrics proto.PsiMetrics

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		// Expecting lines like: "some avg10=... avg60=... avg300=... total=..."
		if len(parts) < 5 {
			continue
		}

		entry := &proto.PsiData{}
		for _, part := range parts[1:] {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				continue
			}
			switch kv[0] {
			case "avg10":
				entry.Avg10, _ = strconv.ParseFloat(kv[1], 64)
			case "avg60":
				entry.Avg60, _ = strconv.ParseFloat(kv[1], 64)
			case "avg300":
				entry.Avg300, _ = strconv.ParseFloat(kv[1], 64)
			case "total":
				entry.Total, _ = strconv.ParseUint(kv[1], 10, 64)
			}
		}

		// Assign parsed PSI data to the appropriate section ("some" or "full")
		switch parts[0] {
		case "some":
			metrics.Some = entry
		case "full":
			metrics.Full = entry
		}
	}

	return &metrics
}
