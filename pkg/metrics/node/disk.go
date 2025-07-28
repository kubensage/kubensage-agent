package node

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
	"strings"
)

func diskIOSummary(counters map[string]disk.IOCountersStat, logger *zap.Logger) *gen.DiskIOSummary {
	logger.Debug("Start diskIOSummary")

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

	summary := &gen.DiskIOSummary{
		TotalReadBytes:  readBytes,
		TotalWriteBytes: writeBytes,
		TotalReadOps:    readOps,
		TotalWriteOps:   writeOps,
	}

	logger.Debug("End diskIOSummary")

	return summary
}

func diskUsages(partitions []disk.PartitionStat, logger *zap.Logger) []*gen.DiskUsage {
	logger.Debug("Start diskUsages")

	diskUsages := make([]*gen.DiskUsage, 0, len(partitions))

	for _, p := range partitions {
		if !isRealFilesystem(p.Fstype, logger) {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		diskUsages = append(diskUsages, &gen.DiskUsage{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	logger.Debug("End diskUsages")

	return diskUsages
}

func isRealFilesystem(fstype string, logger *zap.Logger) bool {
	logger.Debug("Start isRealFilesystem")

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

	logger.Debug("End isRealFilesystem")

	return false
}
