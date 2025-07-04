package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// MemoryMetrics contains memory usage statistics for a container.
// All numeric fields are represented as int64 and default to -1 if the source data is missing.

type MemoryMetrics struct {
	Timestamp       int64 `json:"timestamp,omitempty"`         // Timestamp of the memory stat collection
	WorkingSetBytes int64 `json:"working_set_bytes,omitempty"` // "Working set" memory in bytes
	AvailableBytes  int64 `json:"available_bytes,omitempty"`   // Available memory in bytes
	UsageBytes      int64 `json:"usage_bytes,omitempty"`       // Total memory usage in bytes
	RssBytes        int64 `json:"rss_bytes,omitempty"`         // Resident Set Size (non-swapped) in bytes
	PageFaults      int64 `json:"page_faults,omitempty"`       // Total page faults
	MajorPageFaults int64 `json:"major_page_faults,omitempty"` // Major page faults (disk access)
}

// SafeMemoryMetrics safely extracts memory metrics from a ContainerStats object.
// If stats.Memory is nil, it returns an empty struct with zeroed values.
// Optional fields are safely converted to int64 using a fallback of -1.
func SafeMemoryMetrics(stats *runtimeapi.ContainerStats) MemoryMetrics {
	if stats.Memory == nil {
		return MemoryMetrics{}
	}

	metrics := MemoryMetrics{
		Timestamp:       stats.Memory.Timestamp,
		WorkingSetBytes: utils.SafeUint64ValueToInt64OrDefault(stats.Memory.WorkingSetBytes, -1),
		AvailableBytes:  utils.SafeUint64ValueToInt64OrDefault(stats.Memory.AvailableBytes, -1),
		UsageBytes:      utils.SafeUint64ValueToInt64OrDefault(stats.Memory.UsageBytes, -1),
		RssBytes:        utils.SafeUint64ValueToInt64OrDefault(stats.Memory.RssBytes, -1),
		PageFaults:      utils.SafeUint64ValueToInt64OrDefault(stats.Memory.PageFaults, -1),
		MajorPageFaults: utils.SafeUint64ValueToInt64OrDefault(stats.Memory.MajorPageFaults, -1),
	}

	return metrics
}
