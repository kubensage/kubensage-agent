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
	"sync"
	"time"
)

// Metrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func Metrics(ctx context.Context, interval time.Duration, logger *zap.Logger, topN int) (*gen.NodeMetrics, []error) {
	logger.Debug("Start to collect metrics")

	routines := 19 // Must match the total number of goroutines present below
	var wg sync.WaitGroup
	errChan := make(chan error, routines)
	wg.Add(routines)

	var err error

	// 1
	var info *host.InfoStat
	go func() {
		defer wg.Done()
		logger.Debug("Start host.InfoWithContext")
		info, err = host.InfoWithContext(ctx)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish host.InfoWithContext")
	}()

	// 2
	var cpuInfo []cpu.InfoStat
	go func() {
		logger.Debug("Start cpu.InfoWithContext")
		cpuInfo, err = cpu.InfoWithContext(ctx)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish cpu.InfoWithContext")
	}()

	// 3
	var cpuPercents []float64
	go func() {
		logger.Debug("Start cpu.PercentWithContext")
		cpuPercents, err = cpu.PercentWithContext(ctx, interval, true) // <-- true = per-core
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish cpu.PercentWithContext")
	}()

	// 4
	var totalCpuPercent []float64
	go func() {
		logger.Debug("Start cpu.PercentWithContext")
		totalCpuPercent, err = cpu.PercentWithContext(ctx, interval, false) // <-- true = per-core
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish cpu.PercentWithContext")
	}()

	// 5
	var memInfo *mem.VirtualMemoryStat
	go func() {
		logger.Debug("Start mem.VirtualMemoryWithContext")
		memInfo, err = mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish mem.VirtualMemoryWithContext")
	}()

	// 6
	var netInfoIO []net.IOCountersStat
	go func() {
		logger.Debug("Start net.IOCountersWithContext")
		netInfoIO, err = net.IOCountersWithContext(ctx, false)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish net.IOCountersWithContext")
	}()

	// 7
	var counters map[string]disk.IOCountersStat
	go func() {
		logger.Debug("Start disk.IOCountersWithContext")
		counters, err = disk.IOCountersWithContext(ctx)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish disk.IOCountersWithContext")
	}()

	// 8
	var partitions []disk.PartitionStat
	go func() {
		logger.Debug("Start disk.PartitionsWithContext")
		partitions, err = disk.PartitionsWithContext(ctx, true)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish disk.PartitionsWithContext")
	}()

	// 9
	var interfaces []net.InterfaceStat
	go func() {
		logger.Debug("Start net.InterfacesWithContext")
		interfaces, err = net.InterfacesWithContext(ctx)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish net.InterfacesWithContext")
	}()

	// 10
	var processesMemInfo []*gen.ProcessMemInfo
	go func() {
		logger.Debug("Start ")
		processesMemInfo, err = topMem(ctx, topN, logger)
		if err != nil {
			errChan <- err
		}
		logger.Debug("Finish ")
	}()

	// 11
	var _cpuInfos []*gen.CpuInfo
	go func() {
		_cpuInfos = cpuInfos(cpuInfo, cpuPercents, logger)
	}()

	// 12
	var _netUsage *gen.NetUsage
	go func() {
		_netUsage = netUsage(netInfoIO[0], logger)
	}()

	// 13
	var _diskUsages []*gen.DiskUsage
	go func() {
		_diskUsages = diskUsages(partitions, logger)
	}()

	//14
	var _diskIoSummary *gen.DiskIOSummary
	go func() {
		_diskIoSummary = diskIOSummary(counters, logger)
	}()

	//15
	var _networkInterfaces []*gen.InterfaceStat
	go func() {
		_networkInterfaces = networkInterfaces(interfaces, logger)
	}()

	// 16
	var ipv4, ipv6 string
	go func() {
		ipv4, ipv6 = getPrimaryIPs(_networkInterfaces, logger)
	}()

	// 17
	var cpuPsi, memPsi, ioPsi *gen.PsiMetrics
	go func() {
		cpuPsi = psiMetrics("/proc/pressure/cpu", logger)
	}()

	// 18
	go func() {
		memPsi = psiMetrics("/proc/pressure/memory", logger)
	}()

	// 19
	go func() {
		ioPsi = psiMetrics("/proc/pressure/io", logger)
	}()

	wg.Wait()
	close(errChan)

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

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return nodeInfo, errs
}
