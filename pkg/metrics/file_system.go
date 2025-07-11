package metrics

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeFileSystemMetrics safely extracts file system metrics from a ContainerStats object.
// If the WritableLayer or any subfield is missing, default values are used.
// This function prevents panics when fields are nil and ensures consistency across data collection.
func SafeFileSystemMetrics(stats *runtimeapi.ContainerStats) *proto.FileSystemMetrics {
	if stats.WritableLayer == nil {
		return &proto.FileSystemMetrics{}
	}

	metrics := &proto.FileSystemMetrics{
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
