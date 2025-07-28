package node

import (
	"context"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"time"
)

// Metrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func Metrics(ctx context.Context, interval time.Duration, logger *zap.Logger, topN int) (*gen.NodeMetrics, error) {
	logger.Debug("Start to collect metrics")

	logger.Debug("Start host.InfoWithContext")
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish host.InfoWithContext")

	logger.Debug("Start cpu.InfoWithContext")
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish cpu.InfoWithContext")

	logger.Debug("Start cpu.PercentWithContext")
	cpuPercents, err := cpu.PercentWithContext(ctx, interval, true) // <-- true = per-core
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish cpu.PercentWithContext")

	logger.Debug("Start cpu.PercentWithContext")
	totalCpuPercent, err := cpu.PercentWithContext(ctx, interval, false) // <-- true = per-core
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish cpu.PercentWithContext")

	logger.Debug("Start mem.VirtualMemoryWithContext")
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish mem.VirtualMemoryWithContext")

	logger.Debug("Start net.IOCountersWithContext")
	netInfoIO, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish net.IOCountersWithContext")

	logger.Debug("Start disk.IOCountersWithContext")
	counters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish disk.IOCountersWithContext")

	logger.Debug("Start disk.PartitionsWithContext")
	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish disk.PartitionsWithContext")

	logger.Debug("Start net.InterfacesWithContext")
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish net.InterfacesWithContext")

	logger.Debug("Start ")
	processesMemInfo, err := topMem(ctx, topN, logger)
	if err != nil {
		return nil, err
	}
	logger.Debug("Finish ")

	cpuInfos := cpuInfos(cpuInfo, cpuPercents, logger)
	netUsage := netUsage(netInfoIO[0], logger)
	diskUsages := diskUsages(partitions, logger)
	diskIoSummary := diskIOSummary(counters, logger)
	networkInterfaces := networkInterfaces(interfaces, logger)

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

		ProcessesMemInfo: processesMemInfo,

		DiskUsages:    diskUsages,
		DiskIoSummary: diskIoSummary,

		PsiCpuMetrics:    psiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: psiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     psiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: networkInterfaces,
	}

	ipv4, ipv6 := getPrimaryIPs(networkInterfaces, logger)

	if ipv4 != "" {
		nodeInfo.PrimaryIpv4 = wrapperspb.String(ipv4)
	}
	if ipv6 != "" {
		nodeInfo.PrimaryIpv6 = wrapperspb.String(ipv6)
	}

	logger.Debug("Finish to collect metrics")

	return nodeInfo, nil
}
