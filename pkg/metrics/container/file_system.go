package container

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// fileSystemMetrics safely extracts file system metrics from a ContainerStats object.
// If the WritableLayer or any subfield is missing, default values are used.
// This function prevents panics when fields are nil and ensures consistency across data collection.
func fileSystemMetrics(stats *cri.ContainerStats) *gen.FileSystemMetrics {
	if stats.WritableLayer == nil {
		return &gen.FileSystemMetrics{}
	}

	var usedBytes, inodesUsed *wrapperspb.UInt64Value

	if stats.WritableLayer.UsedBytes != nil {
		usedBytes = utils.ConvertCRIUInt64(stats.WritableLayer.UsedBytes)
	}

	if stats.WritableLayer.InodesUsed != nil {
		inodesUsed = utils.ConvertCRIUInt64(stats.WritableLayer.InodesUsed)
	}

	metrics := &gen.FileSystemMetrics{
		Timestamp:  stats.WritableLayer.Timestamp,
		UsedBytes:  usedBytes,
		InodesUsed: inodesUsed,
	}

	if stats.WritableLayer.FsId != nil {
		metrics.Mountpoint = stats.WritableLayer.FsId.Mountpoint
	} else {
		metrics.Mountpoint = ""
	}

	return metrics
}
