package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SwapMetrics contains swap usage statistics for a container.
// All fields default to -1 if the underlying data is missing or unavailable.
type SwapMetrics struct {
	Timestamp          int64 `json:"timestamp,omitempty"`            // Timestamp of swap stat
	SwapAvailableBytes int64 `json:"swap_available_bytes,omitempty"` // Available swap in bytes
	SwapUsageBytes     int64 `json:"swap_usage_bytes,omitempty"`     // Swap used in bytes
}

// SafeSwapMetrics safely extracts swap-related metrics from a ContainerStats object.
// If the swap field is nil, it returns a zero-value SwapMetrics struct.
// Missing numeric fields are safely converted using default fallback values (typically -1).
func SafeSwapMetrics(stats *runtimeapi.ContainerStats) SwapMetrics {
	if stats.Swap == nil {
		return SwapMetrics{}
	}

	metrics := SwapMetrics{
		Timestamp:          stats.Swap.Timestamp,
		SwapAvailableBytes: utils.SafeUint64ValueToInt64OrDefault(stats.Swap.SwapAvailableBytes, -1),
		SwapUsageBytes:     utils.SafeUint64ValueToInt64OrDefault(stats.Swap.SwapUsageBytes, -1),
	}

	return metrics
}
