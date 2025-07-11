package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeSwapMetrics safely extracts swap-related metrics from a ContainerStats object.
// If the swap field is nil, it returns a zero-value SwapMetrics struct.
// Missing numeric fields are safely converted using default fallback values (typically -1).
func SafeSwapMetrics(stats *runtimeapi.ContainerStats) *proto.SwapMetrics {
	if stats.Swap == nil {
		return &proto.SwapMetrics{}
	}

	metrics := &proto.SwapMetrics{
		Timestamp:      stats.Swap.Timestamp,
		AvailableBytes: utils.SafeUint64ValueToInt64OrDefault(stats.Swap.SwapAvailableBytes, -1),
		UsageBytes:     utils.SafeUint64ValueToInt64OrDefault(stats.Swap.SwapUsageBytes, -1),
	}

	return metrics
}
