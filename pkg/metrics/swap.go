package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type SwapMetrics struct {
	Timestamp          int64 `json:"timestamp,omitempty"`
	SwapAvailableBytes int64 `json:"swap_available_bytes,omitempty"`
	SwapUsageBytes     int64 `json:"swap_usage_bytes,omitempty"`
}

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
