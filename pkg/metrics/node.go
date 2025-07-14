package metrics

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	proto "gitlab.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"strings"
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

	cpuInfos := getCpuInfos(cpuInfo, cpuPercents)
	netUsage := getNetUsage(netInfoIO[0])
	logger.Info("Net len", zap.Int("len", len(netInfoIO)))
	diskUsages := getDiskUsages(partitions)
	diskIoSummary := getDiskIOSummary(counters)
	networkInterfaces := getNetworkInterfaces(interfaces)

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

		NetUsage: netUsage,

		DiskUsages:    diskUsages,
		DiskIoSummary: diskIoSummary,

		PsiCpuMetrics:    getPsiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: getPsiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     getPsiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: networkInterfaces,
	}

	return nodeInfo, nil
}

func getNetUsage(stat net.IOCountersStat) *proto.NetUsage {
	return &proto.NetUsage{
		TotalBytesSent:       stat.BytesSent,
		TotalBytesReceived:   stat.BytesRecv,
		TotalPacketsSent:     stat.PacketsSent,
		TotalPacketsReceived: stat.PacketsRecv,
		TotalErrIn:           stat.Errin,
		TotalErrOut:          stat.Errout,
		TotalDropIn:          stat.Dropin,
		TotalDropOut:         stat.Dropout,
		TotalFifoErrIn:       stat.Fifoin,
		TotalFifoErrOut:      stat.Fifoout,
	}
}

func getCpuInfos(cpuInfo []cpu.InfoStat, cpuPercents []float64) []*proto.CpuInfo {
	var cpuInfos []*proto.CpuInfo
	minLen := len(cpuInfo)
	if len(cpuPercents) < minLen {
		minLen = len(cpuPercents)
	}

	for i := 0; i < minLen; i++ {
		ci := cpuInfo[i]
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
	return cpuInfos
}

func getNetworkInterfaces(interfaces net.InterfaceStatList) []*proto.InterfaceStat {
	networkInterfaces := make([]*proto.InterfaceStat, 0, len(interfaces))

	for _, stat := range interfaces {
		addresses := make([]string, 0, len(stat.Addrs))
		for _, addr := range stat.Addrs {
			addresses = append(addresses, addr.Addr)
		}

		networkInterfaces = append(networkInterfaces, &proto.InterfaceStat{
			Index:        int32(stat.Index),
			Mtu:          int32(stat.MTU),
			Name:         stat.Name,
			HardwareAddr: stat.HardwareAddr,
			Flags:        stat.Flags,
			Addrs:        addresses,
		})
	}
	return networkInterfaces
}

func getDiskIOSummary(counters map[string]disk.IOCountersStat) *proto.DiskIOSummary {
	var readBytes, writeBytes, readOps, writeOps uint64
	for name, stat := range counters {
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		readBytes += stat.ReadBytes
		writeBytes += stat.WriteBytes
		readOps += stat.ReadCount
		writeOps += stat.WriteCount
	}

	summary := &proto.DiskIOSummary{
		TotalReadBytes:  readBytes,
		TotalWriteBytes: writeBytes,
		TotalReadOps:    readOps,
		TotalWriteOps:   writeOps,
	}
	return summary
}

func getDiskUsages(partitions []disk.PartitionStat) []*proto.DiskUsage {
	diskUsages := make([]*proto.DiskUsage, 0, len(partitions))

	for _, p := range partitions {
		if !isRealFilesystem(p.Fstype) {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		diskUsages = append(diskUsages, &proto.DiskUsage{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	return diskUsages
}

func isRealFilesystem(fstype string) bool {
	realFS := []string{
		// Linux
		"ext4", "ext3", "ext2", "xfs", "btrfs", "zfs", "f2fs", "nilfs2",

		// Windows
		"ntfs", "exfat", "fat32", "vfat", "fat", "refs",

		// macOS / Unix
		"apfs", "hfs", "hfs+", "ufs", "ffs",

		// BSD
		"hammer", "hammer2", "zfs", "ufs2",

		// Network / Cluster FS
		"nfs", "nfs4", "cifs", "smbfs", "glusterfs", "ceph", "lustre", "ocfs2", "gfs2",
	}

	for _, fs := range realFS {
		if strings.EqualFold(fstype, fs) {
			return true
		}
	}
	return false
}
