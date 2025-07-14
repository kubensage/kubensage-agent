package container

import (
	"gitlab.com/kubensage/kubensage-agent/pkg/utils"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// swapMetrics safely extracts swap-related metrics from a ContainerStats object.
// If the swap field is nil, it returns a zero-value swapMetrics struct.
// Missing numeric fields are safely converted using default fallback values (typically -1).
func swapMetrics(stats *cri.ContainerStats) *gen.SwapMetrics {
	if stats.Swap == nil {
		return &gen.SwapMetrics{}
	}

	var availableBytes, usageBytes *wrapperspb.UInt64Value

	if stats.Swap.SwapAvailableBytes != nil {
		availableBytes = utils.ConvertCRIUInt64(stats.Swap.SwapAvailableBytes)
	}

	if stats.Swap.SwapUsageBytes != nil {
		usageBytes = utils.ConvertCRIUInt64(stats.Swap.SwapUsageBytes)
	}

	metrics := &gen.SwapMetrics{
		Timestamp:      stats.Swap.Timestamp,
		AvailableBytes: availableBytes,
		UsageBytes:     usageBytes,
	}

	return metrics
}
