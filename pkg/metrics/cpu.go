package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeCpuMetrics safely extracts CPU metrics from a ContainerStats object.
// If the CPU field is nil, it returns an empty CpuMetrics struct.
// For optional numeric values, it returns -1 if the field is missing.
// This function ensures safe access to optional protobuf fields.
func getCpuMetrics(stats *cri.ContainerStats) *proto.CpuMetrics {
	if stats.Cpu == nil {
		return &proto.CpuMetrics{}
	}

	var usageCoreNanoSeconds, usageNanoCores *wrapperspb.UInt64Value

	if stats.Cpu.UsageCoreNanoSeconds != nil {
		usageCoreNanoSeconds = utils.ConvertCRIUInt64(stats.Cpu.UsageCoreNanoSeconds)
	}

	if stats.Cpu.UsageNanoCores != nil {
		usageNanoCores = utils.ConvertCRIUInt64(stats.Cpu.UsageNanoCores)
	}

	metrics := &proto.CpuMetrics{
		Timestamp:            stats.Cpu.Timestamp,
		UsageCoreNanoSeconds: usageCoreNanoSeconds,
		UsageNanoCores:       usageNanoCores,
	}

	return metrics
}
