package metrics

import (
	"context"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"time"
)

// SafeNodeMetrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func getNodeMetrics(ctx context.Context, interval time.Duration, logger *zap.Logger) (*proto.NodeMetrics, error) {
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

	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	networkInterfaces := make([]*proto.InterfaceStat, len(interfaces))

	for _, stat := range interfaces {
		addresses := make([]string, len(stat.Addrs))

		for _, addr := range stat.Addrs {
			addresses = append(addresses, addr.Addr)
		}

		interfaceStat := &proto.InterfaceStat{
			Index:        int32(stat.Index),
			Mtu:          int32(stat.MTU),
			Name:         stat.Name,
			HardwareAddr: stat.HardwareAddr,
			Flags:        stat.Flags,
			Addrs:        addresses,
		}
		networkInterfaces = append(networkInterfaces, interfaceStat)
	}

	var cpuInfos []*proto.CpuInfo
	for i, ci := range cpuInfo {
		cpuInfos = append(cpuInfos, &proto.CpuInfo{
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

	nodeInfo := &proto.NodeMetrics{
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

		PsiCpuMetrics:    getPsiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: getPsiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     getPsiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: networkInterfaces,
	}

	return nodeInfo, nil
}
