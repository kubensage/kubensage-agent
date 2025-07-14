package node

import (
	"bufio"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"os"
	"strconv"
	"strings"
)

// SafePsiMetrics reads and parses pressure stall information (PSI) from the given /proc/pressure/<resource> file.
// It safely extracts "some" and "full" pressure lines and their average/total values.
// If the file cannot be opened or parsed, it logs the error and returns an empty PsiMetrics struct.
func psiMetrics(path string, logger *zap.Logger) *gen.PsiMetrics {
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open metrics file", zap.String("path", path), zap.Error(err))
		return &gen.PsiMetrics{}
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			logger.Error("Error closing PSI file", zap.String("path", path), zap.Error(err))
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

	return &metrics
}
