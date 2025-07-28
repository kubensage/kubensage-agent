package node

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"go.uber.org/zap"
)

func cpuInfos(cpuInfo []cpu.InfoStat, cpuPercents []float64, logger *zap.Logger) []*gen.CpuInfo {
	logger.Debug("Start cpuInfos")

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

	logger.Debug("End cpuInfos")

	return cpuInfos
}
