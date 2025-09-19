package node

import (
	"context"
	"sync"
	"time"

	"github.com/kubensage/go-common/go"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
//
// If one or more collectors fail, partial results are still returned.
// All collectors are safe and tolerant to failures; they append errors instead of panicking.
func BuildNodeMetrics(
	ctx context.Context,
	interval time.Duration,
	logger *zap.Logger,
	topN int,
) (*gen.NodeMetrics, []error, time.Duration) {
	start := time.Now()

	// Durations (RPC/system calls)
	var hostInfoWithContextDuration time.Duration
	var cpuInfoWithContextDuration time.Duration
	var cpuPercentWithContextDuration time.Duration
	var totalCpuPercentWithContextDuration time.Duration
	var virtualMemoryWithContextDuration time.Duration
	var netIOCountersWithContextDuration time.Duration
	var diskIOCountersWithContextDuration time.Duration
	var diskPartitionsWithContextDuration time.Duration
	var netInterfacesWithContextDuration time.Duration

	// Durations (post-processing/build)
	var listCpuInfosDuration time.Duration
	var buildNetUsageDuration time.Duration
	var buildDiskIOSummaryDuration time.Duration
	var listDiskUsagesDuration time.Duration
	var listTopMemDuration time.Duration
	var listNetworkInterfacesDuration time.Duration
	var buildCpuPsiMetricsDuration time.Duration
	var buildMemPsiMetricsDuration time.Duration
	var buildIOPsiMetricsDuration time.Duration

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	addErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		errs = append(errs, e)
		mu.Unlock()
	}

	var info *host.InfoStat
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		info, err = host.InfoWithContext(ctx)
		hostInfoWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		}
	})

	var cpuInfo []cpu.InfoStat
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		cpuInfo, err = cpu.InfoWithContext(ctx)
		cpuInfoWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		}
	})

	var cpuPercents []float64
	var _cpuInfos []*gen.CpuInfo
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		cpuPercents, err = cpu.PercentWithContext(ctx, interval, true)
		cpuPercentWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		} else {
			_cpuInfos, listCpuInfosDuration = listCpuInfos(cpuInfo, cpuPercents)
		}
	})

	var totalCpuPercent []float64
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		totalCpuPercent, err = cpu.PercentWithContext(ctx, interval, false)
		totalCpuPercentWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		}
	})

	var memInfo *mem.VirtualMemoryStat
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		memInfo, err = mem.VirtualMemoryWithContext(ctx)
		virtualMemoryWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		}
	})

	var netInfoIO []net.IOCountersStat
	var _netUsage *gen.NetUsage
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		netInfoIO, err = net.IOCountersWithContext(ctx, false)
		netIOCountersWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		} else {
			_netUsage, buildNetUsageDuration = buildNetUsage(netInfoIO[0])
		}
	})

	var counters map[string]disk.IOCountersStat
	var _diskIoSummary *gen.DiskIOSummary
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		counters, err = disk.IOCountersWithContext(ctx)
		diskIOCountersWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		} else {
			_diskIoSummary, buildDiskIOSummaryDuration = buildDiskIOSummary(counters)
		}
	})

	var partitions []disk.PartitionStat
	var _diskUsages []*gen.DiskUsage
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		partitions, err = disk.PartitionsWithContext(ctx, true)
		diskPartitionsWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		} else {
			_diskUsages, listDiskUsagesDuration = listDiskUsages(partitions)
		}
	})

	var processesMemInfo []*gen.ProcessMemInfo
	gogo.SafeGo(&wg, func() {
		var err error
		processesMemInfo, err, listTopMemDuration = listTopMem(ctx, topN)

		if err != nil {
			addErr(err)
		}
	})

	var interfaces []net.InterfaceStat
	var _networkInterfaces []*gen.InterfaceStat
	var ipv4, ipv6 string
	gogo.SafeGo(&wg, func() {
		start := time.Now()

		var err error
		interfaces, err = net.InterfacesWithContext(ctx)
		netInterfacesWithContextDuration = time.Since(start)

		if err != nil {
			addErr(err)
		}

		_networkInterfaces, listNetworkInterfacesDuration = listNetworkInterfaces(interfaces)
		ipv4, ipv6 = buildPrimaryIPs(_networkInterfaces)
	})

	var cpuPsi, memPsi, ioPsi *gen.PsiMetrics
	gogo.SafeGo(&wg, func() {
		cpuPsi, buildCpuPsiMetricsDuration = buildPsiMetrics("/proc/pressure/cpu", logger)
	})

	gogo.SafeGo(&wg, func() {
		memPsi, buildMemPsiMetricsDuration = buildPsiMetrics("/proc/pressure/memory", logger)
	})

	gogo.SafeGo(&wg, func() {
		ioPsi, buildIOPsiMetricsDuration = buildPsiMetrics("/proc/pressure/io", logger)
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

	total := time.Since(start)

	logger.Debug("node metrics durations",
		zap.Duration("host_info", hostInfoWithContextDuration),
		zap.Duration("cpu_info", cpuInfoWithContextDuration),
		zap.Duration("cpu_percent_percpu", cpuPercentWithContextDuration),
		zap.Duration("cpu_percent_total", totalCpuPercentWithContextDuration),
		zap.Duration("virtual_memory", virtualMemoryWithContextDuration),
		zap.Duration("net_iocounters", netIOCountersWithContextDuration),
		zap.Duration("disk_iocounters", diskIOCountersWithContextDuration),
		zap.Duration("disk_partitions", diskPartitionsWithContextDuration),
		zap.Duration("net_interfaces", netInterfacesWithContextDuration),

		zap.Duration("list_cpu_infos", listCpuInfosDuration),
		zap.Duration("build_net_usage", buildNetUsageDuration),
		zap.Duration("build_disk_io_summary", buildDiskIOSummaryDuration),
		zap.Duration("list_disk_usages", listDiskUsagesDuration),
		zap.Duration("list_top_mem", listTopMemDuration),
		zap.Duration("list_network_interfaces", listNetworkInterfacesDuration),
		zap.Duration("build_cpu_psi", buildCpuPsiMetricsDuration),
		zap.Duration("build_mem_psi", buildMemPsiMetricsDuration),
		zap.Duration("build_io_psi", buildIOPsiMetricsDuration),

		zap.Duration("total", total),
	)

	return nodeInfo, errs, time.Since(start)
}
