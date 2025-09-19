package container

import (
	"time"

	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// buildSwapMetrics constructs a *gen.SwapMetrics message from the Swap section
// of a CRI ContainerStats object.
//
// This function safely extracts optional swap memory metrics. If the Swap section
// is missing (nil), an empty SwapMetrics struct is returned.
// Values such as swap available and swap usage are converted to
// protobuf UInt64Value wrappers if present.
//
// Parameters:
//   - stats: *cri.ContainerStats
//     The container statistics provided by the CRI runtime, which may include
//     swap usage details under the `Swap` field.
//
// Returns:
//
//   - *gen.SwapMetrics: A SwapMetrics object with the following fields:
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func buildSwapMetrics(
	stats *cri.ContainerStats,
) (*gen.SwapMetrics, time.Duration) {
	start := time.Now()

	if stats.Swap == nil {
		return &gen.SwapMetrics{}, time.Since(start)
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

	return metrics, time.Since(start)
}
