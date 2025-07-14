package container

import (
	"gitlab.com/kubensage/kubensage-agent/pkg/utils"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// cpuMetrics safely extracts CPU metrics from a ContainerStats object.
// If the CPU field is nil, it returns an empty cpuMetrics struct.
// For optional numeric values, it returns -1 if the field is missing.
// This function ensures safe access to optional protobuf fields.
func cpuMetrics(stats *cri.ContainerStats) *gen.CpuMetrics {
	if stats.Cpu == nil {
		return &gen.CpuMetrics{}
	}

	var usageCoreNanoSeconds, usageNanoCores *wrapperspb.UInt64Value

	if stats.Cpu.UsageCoreNanoSeconds != nil {
		usageCoreNanoSeconds = utils.ConvertCRIUInt64(stats.Cpu.UsageCoreNanoSeconds)
	}

	if stats.Cpu.UsageNanoCores != nil {
		usageNanoCores = utils.ConvertCRIUInt64(stats.Cpu.UsageNanoCores)
	}

	metrics := &gen.CpuMetrics{
		Timestamp:            stats.Cpu.Timestamp,
		UsageCoreNanoSeconds: usageCoreNanoSeconds,
		UsageNanoCores:       usageNanoCores,
	}

	return metrics
}
