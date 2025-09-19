package container

import (
	"time"

	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// buildMemoryMetrics constructs a *gen.MemoryMetrics message from the Memory section
// of a CRI ContainerStats object.
//
// This function safely handles missing or optional fields. If the Memory section is nil,
// it returns an empty MemoryMetrics struct. All memory-related counters are converted to
// protobuf UInt64Value wrappers only if present.
//
// Parameters:
//   - stats: *cri.ContainerStats
//     The container statistics provided by the CRI runtime, including memory usage metrics.
//
// Returns:
//   - *gen.MemoryMetrics: A populated MemoryMetrics object with the following fields:
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func buildMemoryMetrics(
	stats *cri.ContainerStats,
) (*gen.MemoryMetrics, time.Duration) {
	start := time.Now()

	if stats.Memory == nil {
		return &gen.MemoryMetrics{}, time.Since(start)
	}

	var workingSetBytes, availableBytes, usageBytes, rssBytes, pageFaults, majorPageFaults *wrapperspb.UInt64Value

	if stats.Memory.WorkingSetBytes != nil {
		workingSetBytes = utils.ConvertCRIUInt64(stats.Memory.WorkingSetBytes)
	}

	if stats.Memory.AvailableBytes != nil {
		availableBytes = utils.ConvertCRIUInt64(stats.Memory.AvailableBytes)
	}

	if stats.Memory.UsageBytes != nil {
		usageBytes = utils.ConvertCRIUInt64(stats.Memory.UsageBytes)
	}

	if stats.Memory.RssBytes != nil {
		rssBytes = utils.ConvertCRIUInt64(stats.Memory.RssBytes)
	}

	if stats.Memory.PageFaults != nil {
		pageFaults = utils.ConvertCRIUInt64(stats.Memory.PageFaults)
	}

	if stats.Memory.MajorPageFaults != nil {
		majorPageFaults = utils.ConvertCRIUInt64(stats.Memory.MajorPageFaults)
	}

	metrics := &gen.MemoryMetrics{
		Timestamp:       stats.Memory.Timestamp,
		WorkingSetBytes: workingSetBytes,
		AvailableBytes:  availableBytes,
		UsageBytes:      usageBytes,
		RssBytes:        rssBytes,
		PageFaults:      pageFaults,
		MajorPageFaults: majorPageFaults,
	}

	return metrics, time.Since(start)
}
