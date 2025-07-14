package node

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
)

func cpuInfos(cpuInfo []cpu.InfoStat, cpuPercents []float64) []*gen.CpuInfo {
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
	return cpuInfos
}
