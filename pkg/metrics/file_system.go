package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// FileSystemMetrics represents usage statistics for a container's writable layer.
// All fields are safe defaults: numerical fields fallback to -1 if missing.
type FileSystemMetrics struct {
	Timestamp  int64  // Timestamp of the filesystem stat
	Mountpoint string // Filesystem mount path
	UsedBytes  int64  // Space used in bytes
	InodesUsed int64  // Number of inodes used
}

// SafeFileSystemMetrics safely extracts file system metrics from a ContainerStats object.
// If the WritableLayer or any subfield is missing, default values are used.
// This function prevents panics when fields are nil and ensures consistency across data collection.
func SafeFileSystemMetrics(stats *runtimeapi.ContainerStats) FileSystemMetrics {
	if stats.WritableLayer == nil {
		return FileSystemMetrics{}
	}

	metrics := FileSystemMetrics{
		Timestamp:  stats.WritableLayer.Timestamp,
		UsedBytes:  utils.SafeUint64ValueToInt64OrDefault(stats.WritableLayer.UsedBytes, -1),
		InodesUsed: utils.SafeUint64ValueToInt64OrDefault(stats.WritableLayer.InodesUsed, -1),
	}

	if stats.WritableLayer.FsId != nil {
		metrics.Mountpoint = stats.WritableLayer.FsId.Mountpoint
	} else {
		metrics.Mountpoint = ""
	}

	return metrics
}
