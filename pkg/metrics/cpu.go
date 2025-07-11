package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeCpuMetrics safely extracts CPU metrics from a ContainerStats object.
// If the CPU field is nil, it returns an empty CpuMetrics struct.
// For optional numeric values, it returns -1 if the field is missing.
// This function ensures safe access to optional protobuf fields.
func SafeCpuMetrics(stats *runtimeapi.ContainerStats) *proto.CpuMetrics {
	if stats.Cpu == nil {
		return &proto.CpuMetrics{}
	}

	metrics := &proto.CpuMetrics{
		Timestamp:            stats.Cpu.Timestamp,
		UsageCoreNanoSeconds: utils.SafeUint64ValueToInt64OrDefault(stats.Cpu.UsageCoreNanoSeconds, -1),
		UsageNanoCores:       utils.SafeUint64ValueToInt64OrDefault(stats.Cpu.UsageNanoCores, -1),
	}

	return metrics
}
