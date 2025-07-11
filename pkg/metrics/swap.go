package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeSwapMetrics safely extracts swap-related metrics from a ContainerStats object.
// If the swap field is nil, it returns a zero-value SwapMetrics struct.
// Missing numeric fields are safely converted using default fallback values (typically -1).
func SafeSwapMetrics(stats *runtimeapi.ContainerStats) *proto.SwapMetrics {
	if stats.Swap == nil {
		return &proto.SwapMetrics{}
	}

	var availableBytes, usageBytes *wrapperspb.UInt64Value

	if stats.Swap.SwapAvailableBytes != nil {
		availableBytes = utils.ConvertCRIUInt64(stats.Swap.SwapAvailableBytes)
	}

	if stats.Swap.SwapUsageBytes != nil {
		usageBytes = utils.ConvertCRIUInt64(stats.Swap.SwapUsageBytes)
	}

	metrics := &proto.SwapMetrics{
		Timestamp:      stats.Swap.Timestamp,
		AvailableBytes: availableBytes,
		UsageBytes:     usageBytes,
	}

	return metrics
}
