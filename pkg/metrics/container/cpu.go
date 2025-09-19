package container

import (
	"time"

	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// buildCpuMetrics constructs a *gen.CpuMetrics message from a given CRI ContainerStats.
//
// This function safely handles optional CPU metrics fields by checking for nil pointers
// and wrapping them with protobuf-compatible wrappers. If the input stats do not contain
// CPU data, it returns an empty CpuMetrics object.
//
// Parameters:
//   - stats: *cri.ContainerStats
//     The container statistics object from the CRI runtime. Expected to include CPU usage data.
//
// Returns:
//   - *gen.CpuMetrics: A populated CpuMetrics object with:
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func buildCpuMetrics(
	stats *cri.ContainerStats,
) (*gen.CpuMetrics, time.Duration) {
	start := time.Now()

	if stats.Cpu == nil {
		return &gen.CpuMetrics{}, time.Since(start)
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

	return metrics, time.Since(start)
}
