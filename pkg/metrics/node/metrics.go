package node

import (
	"context"
	"github.com/kubensage/go-common/go"
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

// BuildNodeMetrics collects system-level metrics from the node,
// using gopsutil for hardware and OS information and /proc/pressure
// for PSI metrics (Linux-specific).
//
// The function runs multiple metric collectors concurrently to improve efficiency.
// It gathers information including:
//
//   - Host info (OS, kernel, uptime, hostname, etc.)
//   - CPU info (per-core and total usage)
//   - Memory usage
//   - Network usage
//   - Disk I/O and disk usage
//   - PSI (Pressure Stall Information) for CPU, memory, and IO
//   - Network interfaces and primary IP addresses
//   - Top N memory-consuming processes
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - interval: Duration for calculating CPU usage percentages
//   - logger: Structured logger for debug/info/error messages
//   - topN: Number of processes to include in top memory usage
//
// Returns:
//   - *gen.NodeMetrics: Complete set of collected node-level metrics
//   - []error: List of non-fatal errors encountered during metric collection
//
// If one or more collectors fail, partial results are still returned.
// All collectors are safe and tolerant to failures; they append errors instead of panicking.
func BuildNodeMetrics(
	ctx context.Context,
	interval time.Duration,
	logger *zap.Logger,
	topN int,
) (*gen.NodeMetrics, []error) {
	logger.Debug("Start to collect metrics")

	var wg sync.WaitGroup
	var errs []error

	var info *host.InfoStat
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start host.InfoWithContext")
		info, err = host.InfoWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish host.InfoWithContext")
	})

	var cpuInfo []cpu.InfoStat
	gogo.SafeGo(&wg, func() {
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
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start cpu.PercentWithContext")
		cpuPercents, err = cpu.PercentWithContext(ctx, interval, true)
		if err != nil {
			errs = append(errs, err)
		} else {
			_cpuInfos = listCpuInfos(cpuInfo, cpuPercents, logger)
		}
		logger.Debug("Finish cpu.PercentWithContext")
	})

	var totalCpuPercent []float64
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start cpu.PercentWithContext")
		totalCpuPercent, err = cpu.PercentWithContext(ctx, interval, false)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish cpu.PercentWithContext")
	})

	var memInfo *mem.VirtualMemoryStat
	gogo.SafeGo(&wg, func() {
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
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start net.IOCountersWithContext")
		netInfoIO, err = net.IOCountersWithContext(ctx, false)
		if err != nil {
			errs = append(errs, err)
		} else {
			_netUsage = buildNetUsage(netInfoIO[0], logger)
		}
		logger.Debug("Finish net.IOCountersWithContext")
	})

	var counters map[string]disk.IOCountersStat
	var _diskIoSummary *gen.DiskIOSummary
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start disk.IOCountersWithContext")
		counters, err = disk.IOCountersWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		} else {
			_diskIoSummary = buildDiskIOSummary(counters, logger)
		}
		logger.Debug("Finish disk.IOCountersWithContext")
	})

	var partitions []disk.PartitionStat
	var _diskUsages []*gen.DiskUsage
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start disk.PartitionsWithContext")
		partitions, err = disk.PartitionsWithContext(ctx, true)
		if err != nil {
			errs = append(errs, err)
		} else {
			_diskUsages = ListDiskUsages(partitions, logger)
		}
		logger.Debug("Finish disk.PartitionsWithContext")
	})

	var processesMemInfo []*gen.ProcessMemInfo
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start ")
		processesMemInfo, err = listTopMem(ctx, topN, logger)
		if err != nil {
			errs = append(errs, err)
		}
		logger.Debug("Finish ")
	})

	var interfaces []net.InterfaceStat
	var _networkInterfaces []*gen.InterfaceStat
	var ipv4, ipv6 string
	gogo.SafeGo(&wg, func() {
		var err error
		logger.Debug("Start net.InterfacesWithContext")
		interfaces, err = net.InterfacesWithContext(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		_networkInterfaces = listNetworkInterfaces(interfaces, logger)
		ipv4, ipv6 = buildPrimaryIPs(_networkInterfaces, logger)
		logger.Debug("Finish net.InterfacesWithContext")
	})

	var cpuPsi, memPsi, ioPsi *gen.PsiMetrics
	gogo.SafeGo(&wg, func() {
		cpuPsi = buildPsiMetrics("/proc/pressure/cpu", logger)
	})

	gogo.SafeGo(&wg, func() {
		memPsi = buildPsiMetrics("/proc/pressure/memory", logger)
	})

	gogo.SafeGo(&wg, func() {
		ioPsi = buildPsiMetrics("/proc/pressure/io", logger)
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
