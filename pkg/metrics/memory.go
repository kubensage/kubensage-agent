package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeMemoryMetrics safely extracts memory metrics from a ContainerStats object.
// If stats.Memory is nil, it returns an empty struct with zeroed values.
// Optional fields are safely converted to int64 using a fallback of -1.
func SafeMemoryMetrics(stats *runtimeapi.ContainerStats) *proto.MemoryMetrics {
	if stats.Memory == nil {
		return &proto.MemoryMetrics{}
	}

	metrics := &proto.MemoryMetrics{
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
