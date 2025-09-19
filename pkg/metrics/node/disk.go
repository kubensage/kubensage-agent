package node

import (
	"strings"
	"time"

	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/disk"
)

// listDiskUsages returns disk usage metrics for each valid, real filesystem mountpoint
// provided in the list of disk partitions.
//
// It filters out virtual or non-persistent filesystems (e.g., tmpfs, procfs) using
// the isRealFilesystem check, and skips partitions where usage data is unavailable
// or total capacity is reported as zero.
//
// Each returned *gen.DiskUsage includes fields such as device name, mountpoint,
// filesystem type, total space, used and free space, and percentage used.
//
// Parameters:
//
//   - partitions []disk.PartitionStat:
//     A slice of partition records, typically retrieved via gopsutil's disk.Partitions().
//
// Returns:
//   - []*gen.DiskUsage: a slice of usage summaries, one for each real filesystem with retrievable stats.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func listDiskUsages(
	partitions []disk.PartitionStat,
) ([]*gen.DiskUsage, time.Duration) {
	start := time.Now()

	diskUsages := make([]*gen.DiskUsage, 0, len(partitions))

	for _, p := range partitions {
		if !isRealFilesystem(p.Fstype) {
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

	return diskUsages, time.Since(start)
}

// buildDiskIOSummary aggregates global disk I/O statistics from a map of per-device counters.
//
// It sums read and write bytes, as well as read and write operation counts, across all devices,
// excluding loop and RAM-based devices (e.g., /dev/loop*, /dev/ram*) which are considered ephemeral
// or not relevant for disk I/O reporting.
//
// Parameters:
//
//   - counters map[string]disk.IOCountersStat:
//     A map of device names to I/O statistics, typically retrieved from gopsutil's disk.IOCountersWithContext().
//
// Returns:
//   - *gen.DiskIOSummary:
//     A protobuf-compatible summary containing total read/write byte counts and operation counts
//     across all relevant block devices.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func buildDiskIOSummary(
	counters map[string]disk.IOCountersStat,
) (*gen.DiskIOSummary, time.Duration) {
	start := time.Now()

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

	return summary, time.Since(start)
}

// isRealFilesystem determines whether a given filesystem type represents a persistent
// or mountable storage backend.
//
// It is primarily used to filter out non-storage or ephemeral filesystems such as
// "tmpfs", "proc", or "sysfs", which should not be included in disk usage metrics.
//
// The function compares the input type against a known list of real filesystems used
// across various platforms, including:
//   - Linux (e.g., ext4, xfs, btrfs)
//   - Windows (e.g., ntfs, fat32)
//   - macOS/Unix (e.g., apfs, hfs)
//   - BSD (e.g., hammer2, zfs)
//   - Network or clustered FS (e.g., nfs, ceph, glusterfs)
//
// The comparison is case-insensitive.
//
// Parameters:
//   - fstype string:
//     Filesystem type string (typically from a mount or partition record)
//
// Returns:
//   - bool:
//     True if the filesystem type is considered persistent/mountable, false otherwise.
func isRealFilesystem(
	fstype string,
) bool {
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
