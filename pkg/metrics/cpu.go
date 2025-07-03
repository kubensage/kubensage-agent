package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type CpuMetrics struct {
	Timestamp            int64 `json:"timestamp,omitempty"`
	UsageCoreNanoSeconds int64 `json:"usage_core_nano_seconds,omitempty"`
	UsageNanoCores       int64 `json:"usage_nano_cores,omitempty"`
}

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
