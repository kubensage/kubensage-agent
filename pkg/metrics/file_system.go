package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type FileSystemMetrics struct {
	Timestamp  int64  `json:"timestamp,omitempty"`
	Mountpoint string `json:"mountpoint,omitempty"`
	UsedBytes  int64  `json:"used_bytes,omitempty"`
	InodesUsed int64  `json:"inodes_used,omitempty"`
}

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
