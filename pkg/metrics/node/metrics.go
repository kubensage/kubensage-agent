package node

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"time"
)

// Metrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func Metrics(ctx context.Context, interval time.Duration, logger *zap.Logger) (*gen.NodeMetrics, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuPercents, err := cpu.PercentWithContext(ctx, interval, true) // <-- true = per-core
	if err != nil {
		return nil, err
	}

	totalCpuPercent, err := cpu.PercentWithContext(ctx, interval, false) // <-- true = per-core
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	netInfoIO, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	counters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, err
	}

	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuInfos := cpuInfos(cpuInfo, cpuPercents)
	netUsage := netUsage(netInfoIO[0])
	diskUsages := diskUsages(partitions)
	diskIoSummary := diskIOSummary(counters)
	networkInterfaces := networkInterfaces(interfaces)

	nodeInfo := &gen.NodeMetrics{
		Hostname: info.Hostname,
		Uptime:   info.Uptime,
		BootTime: info.BootTime,
		Procs:    info.Procs,

		Os:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		HostId:          info.HostID,

		TotalCpuPercentage: totalCpuPercent[0],
		CpuInfos:           cpuInfos,

		TotalMemory:     memInfo.Total,
		AvailableMemory: memInfo.Available,
		UsedMemory:      memInfo.Used,
		MemoryUsedPerc:  memInfo.UsedPercent,

		NetUsage: netUsage,

		DiskUsages:    diskUsages,
		DiskIoSummary: diskIoSummary,

		PsiCpuMetrics:    psiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: psiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     psiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: networkInterfaces,
	}

	return nodeInfo, nil
}
