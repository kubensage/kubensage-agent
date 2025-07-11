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
func SafeNodeMetrics(ctx context.Context, interval time.Duration, logger *zap.Logger) (*proto.NodeMetrics, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get CPU usage over 1 second interval (non-per-core)
	cpuPercent, err := cpu.PercentWithContext(ctx, interval, false)

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

		CpuModel:        cpuInfo[0].ModelName,
		CpuCores:        cpuInfo[0].Cores,
		CpuUsagePercent: cpuPercent[0],

		TotalMemory:    memInfo.Total,
		FreeMemory:     memInfo.Free,
		UsedMemory:     memInfo.Used,
		MemoryUsedPerc: memInfo.UsedPercent,

		PsiCpuMetrics:    SafePsiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: SafePsiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     SafePsiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: networkInterfaces,
	}

	return nodeInfo, nil
}
