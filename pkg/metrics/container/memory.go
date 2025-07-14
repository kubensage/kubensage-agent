package container

import (
	"gitlab.com/kubensage/kubensage-agent/pkg/utils"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// memoryMetrics safely extracts memory metrics from a ContainerStats object.
// If stats.Memory is nil, it returns an empty struct with zeroed values.
// Optional fields are safely converted to int64 using a fallback of -1.
func memoryMetrics(stats *cri.ContainerStats) *gen.MemoryMetrics {
	if stats.Memory == nil {
		return &gen.MemoryMetrics{}
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

	return metrics
}
