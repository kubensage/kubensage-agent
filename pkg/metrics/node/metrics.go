package node

import (
	"context"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"sync"
	"time"
)

// Metrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func Metrics(ctx context.Context, interval time.Duration, logger *zap.Logger, topN int) (*gen.NodeMetrics, []error) {
	logger.Debug("Start to collect metrics")

	var wg sync.WaitGroup
	var errs []error

	var info *host.InfoStat
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start host.InfoWithContext")
		info, err = host.InfoWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish host.InfoWithContext")
	})

	var cpuInfo []cpu.InfoStat
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start cpu.InfoWithContext")
		cpuInfo, err = cpu.InfoWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish cpu.InfoWithContext")
	})

	var cpuPercents []float64
	var _cpuInfos []*gen.CpuInfo
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start cpu.PercentWithContext")
		cpuPercents, err = cpu.PercentWithContext(ctx, interval, true) // <-- true = per-core
		if err != nil {
			errs = append(errs, err)
		} else {
			_cpuInfos = cpuInfos(cpuInfo, cpuPercents, logger)
		}
		logger.Debug("Finish cpu.PercentWithContext")
	})

	var totalCpuPercent []float64
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start cpu.PercentWithContext")
		totalCpuPercent, err = cpu.PercentWithContext(ctx, interval, false) // <-- true = per-core
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish cpu.PercentWithContext")
	})

	var memInfo *mem.VirtualMemoryStat
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start mem.VirtualMemoryWithContext")
		memInfo, err = mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish mem.VirtualMemoryWithContext")
	})

	var netInfoIO []net.IOCountersStat
	var _netUsage *gen.NetUsage
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start net.IOCountersWithContext")
		netInfoIO, err = net.IOCountersWithContext(ctx, false)
		if err != nil {
			errs = append(errs, err)
		} else {
			_netUsage = netUsage(netInfoIO[0], logger)
		}
		logger.Debug("Finish net.IOCountersWithContext")
	})

	var counters map[string]disk.IOCountersStat
	var _diskIoSummary *gen.DiskIOSummary
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start disk.IOCountersWithContext")
		counters, err = disk.IOCountersWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		} else {
			_diskIoSummary = diskIOSummary(counters, logger)
		}
		logger.Debug("Finish disk.IOCountersWithContext")
	})

	var partitions []disk.PartitionStat
	var _diskUsages []*gen.DiskUsage
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start disk.PartitionsWithContext")
		partitions, err = disk.PartitionsWithContext(ctx, true)
		if err != nil {
			errs = append(errs, err)
		} else {
			_diskUsages = diskUsages(partitions, logger)
		}
		logger.Debug("Finish disk.PartitionsWithContext")
	})

	var processesMemInfo []*gen.ProcessMemInfo
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start ")
		processesMemInfo, err = topMem(ctx, topN, logger)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish ")
	})

	var interfaces []net.InterfaceStat
	var _networkInterfaces []*gen.InterfaceStat
	var ipv4, ipv6 string
	utils.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start net.InterfacesWithContext")
		interfaces, err = net.InterfacesWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		_networkInterfaces = networkInterfaces(interfaces, logger)
		ipv4, ipv6 = getPrimaryIPs(_networkInterfaces, logger)
		logger.Debug("Finish net.InterfacesWithContext")
	})

	var cpuPsi, memPsi, ioPsi *gen.PsiMetrics
	utils.SafeGo(&wg, func() {
		cpuPsi = psiMetrics("/proc/pressure/cpu", logger)
	})

	utils.SafeGo(&wg, func() {
		memPsi = psiMetrics("/proc/pressure/memory", logger)
	})

	utils.SafeGo(&wg, func() {
		ioPsi = psiMetrics("/proc/pressure/io", logger)
	})

	wg.Wait()

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
		CpuInfos:           _cpuInfos,

		TotalMemory:     memInfo.Total,
		AvailableMemory: memInfo.Available,
		UsedMemory:      memInfo.Used,
		MemoryUsedPerc:  memInfo.UsedPercent,

		NetUsage: _netUsage,

		ProcessesMemInfo: processesMemInfo,

		DiskUsages:    _diskUsages,
		DiskIoSummary: _diskIoSummary,

		PsiCpuMetrics:    cpuPsi,
		PsiMemoryMetrics: memPsi,
		PsiIoMetrics:     ioPsi,

		NetworkInterfaces: _networkInterfaces,
	}

	if ipv4 != "" {
		nodeInfo.PrimaryIpv4 = wrapperspb.String(ipv4)
	}
	if ipv6 != "" {
		nodeInfo.PrimaryIpv6 = wrapperspb.String(ipv6)
	}

	logger.Debug("Finish to collect metrics")

	return nodeInfo, errs
}
