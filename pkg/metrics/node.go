package metrics

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

func SafeNodeMetrics(ctx context.Context) (*NodeMetrics, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuPercent, err := cpu.PercentWithContext(ctx, 1*time.Second, false)

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	netInfo, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	nodeInfo := &NodeMetrics{
		Hostname:        info.Hostname,
		Uptime:          info.Uptime,
		BootTime:        info.BootTime,
		Procs:           info.Procs,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		HostID:          info.HostID,

		CPUModel:        cpuInfo[0].ModelName,
		CPUCores:        cpuInfo[0].Cores,
		CPUUsagePercent: cpuPercent[0],

		TotalMemory:    memInfo.Total,
		FreeMemory:     memInfo.Free,
		UsedMemory:     memInfo.Used,
		MemoryUsedPerc: memInfo.UsedPercent,

		NetworkInterfaces: netInfo,
	}

	return nodeInfo, nil
}
