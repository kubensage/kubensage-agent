package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type MemoryMetrics struct {
	Timestamp       int64 `json:"timestamp,omitempty"`
	WorkingSetBytes int64 `json:"working_set_bytes,omitempty"`
	AvailableBytes  int64 `json:"available_bytes,omitempty"`
	UsageBytes      int64 `json:"usage_bytes,omitempty"`
	RssBytes        int64 `json:"rss_bytes,omitempty"`
	PageFaults      int64 `json:"page_faults,omitempty"`
	MajorPageFaults int64 `json:"major_page_faults,omitempty"`
}

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
