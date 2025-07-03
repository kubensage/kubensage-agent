package metrics

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

type NodeMetrics struct {
	Hostname        string `json:"hostname,omitempty"`
	Uptime          uint64 `json:"uptime,omitempty"`
	BootTime        uint64 `json:"boot_time,omitempty"`
	Procs           uint64 `json:"procs,omitempty"`
	OS              string `json:"os,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformFamily  string `json:"platform_family,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	KernelVersion   string `json:"kernel_version,omitempty"`
	KernelArch      string `json:"kernel_arch,omitempty"`
	HostID          string `json:"host_id,omitempty"`

	CPUModel        string  `json:"cpu_model,omitempty"`
	CPUCores        int32   `json:"cpu_cores,omitempty"`
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"`

	TotalMemory    uint64  `json:"total_memory,omitempty"`
	FreeMemory     uint64  `json:"free_memory,omitempty"`
	UsedMemory     uint64  `json:"used_memory,omitempty"`
	MemoryUsedPerc float64 `json:"memory_used_perc,omitempty"`

	PsiCpuMetrics    PsiMetrics `json:"psi_cpu_metrics,omitempty"`
	PsiMemoryMetrics PsiMetrics `json:"psi_memory_metrics,omitempty"`
	PsiIoMetrics     PsiMetrics `json:"psi_io_metrics,omitempty"`

	NetworkInterfaces []net.InterfaceStat `json:"network_interfaces,omitempty"`
}

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

		PsiCpuMetrics:    SafePsiMetrics("/proc/pressure/cpu"),
		PsiMemoryMetrics: SafePsiMetrics("/proc/pressure/memory"),
		PsiIoMetrics:     SafePsiMetrics("/proc/pressure/io"),

		NetworkInterfaces: netInfo,
	}

	return nodeInfo, nil
}
