package node

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"go.uber.org/zap"
)

// listCpuInfos combines static CPU metadata and usage percentages into a slice of CpuInfo messages.
//
// It assumes that both input slices are indexed by logical CPU order, and only processes up to the
// minimum of their lengths to avoid panics due to slice size mismatch.
//
// Each resulting *gen.CpuInfo message contains:
//   - CPU index (logical core number)
//   - Model name, vendor ID, physical/core IDs
//   - Clock speed in MHz
//   - Number of physical cores (as reported)
//   - Usage percentage over the sampled interval
//
// Parameters:
//   - cpuInfo []cpu.InfoStat:
//     Slice of CPU metadata structs, typically returned by gopsutil's cpu.InfoWithContext.
//   - cpuPercents []float64:
//     Slice of usage percentages, typically returned by cpu.PercentWithContext with `percpu=true`.
//   - logger *zap.Logger:
//     Structured logger used for debug tracing during the operation.
//
// Returns:
//   - []*gen.CpuInfo:
//     A slice of protobuf CpuInfo messages representing each logical CPU core's metadata and usage.
func listCpuInfos(
	cpuInfo []cpu.InfoStat,
	cpuPercents []float64,
	logger *zap.Logger,
) []*gen.CpuInfo {
	logger.Debug("Start listCpuInfos")

	var cpuInfos []*gen.CpuInfo
	minLen := len(cpuInfo)
	if len(cpuPercents) < minLen {
		minLen = len(cpuPercents)
	}

	for i := 0; i < minLen; i++ {
		ci := cpuInfo[i]
		cpuInfos = append(cpuInfos, &gen.CpuInfo{
			Model:      ci.ModelName,
			Cores:      ci.Cores,
			Mhz:        int32(ci.Mhz),
			VendorId:   ci.VendorID,
			PhysicalId: ci.PhysicalID,
			CoreId:     ci.CoreID,
			Cpu:        ci.CPU,
			Usage:      cpuPercents[i],
		})
	}

	logger.Debug("End listCpuInfos")

	return cpuInfos
}
