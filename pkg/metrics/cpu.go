package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// CpuMetrics holds CPU usage statistics extracted from CRI ContainerStats.
// Values are normalized as int64, and missing fields are represented by -1.
type CpuMetrics struct {
	Timestamp            int64 `json:"timestamp,omitempty"`               // Timestamp of the CPU stat collection
	UsageCoreNanoSeconds int64 `json:"usage_core_nano_seconds,omitempty"` // Cumulative CPU usage in nanoseconds
	UsageNanoCores       int64 `json:"usage_nano_cores,omitempty"`        // Instantaneous CPU usage in nano cores
}

// SafeCpuMetrics safely extracts CPU metrics from a ContainerStats object.
// If the CPU field is nil, it returns an empty CpuMetrics struct.
// For optional numeric values, it returns -1 if the field is missing.
// This function ensures safe access to optional protobuf fields.
func SafeCpuMetrics(stats *runtimeapi.ContainerStats) CpuMetrics {
	if stats.Cpu == nil {
		return CpuMetrics{}
	}

	metrics := CpuMetrics{
		Timestamp:            stats.Cpu.Timestamp,
		UsageCoreNanoSeconds: utils.SafeUint64ValueToInt64OrDefault(stats.Cpu.UsageCoreNanoSeconds, -1),
		UsageNanoCores:       utils.SafeUint64ValueToInt64OrDefault(stats.Cpu.UsageNanoCores, -1),
	}

	return metrics
}
