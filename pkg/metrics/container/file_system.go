package container

import (
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// buildFileSystemMetrics constructs a *gen.FileSystemMetrics message from the
// WritableLayer section of a given CRI ContainerStats object.
//
// This function handles missing or optional fields gracefully. If WritableLayer
// is nil, it returns an empty FileSystemMetrics struct. Optional numeric fields
// (e.g., UsedBytes, InodesUsed) are converted to protobuf wrapper types only if present.
// The mountpoint is extracted from FsId.Mountpoint, or set to an empty string if unavailable.
//
// Parameters:
//   - stats: *cri.ContainerStats
//     The container statistics from the CRI runtime, expected to include filesystem usage data.
//
// Returns:
//   - *gen.FileSystemMetrics
//     A populated FileSystemMetrics object with:
//   - Timestamp: from WritableLayer.Timestamp
//   - UsedBytes: wrapped value from WritableLayer.UsedBytes (optional)
//   - InodesUsed: wrapped value from WritableLayer.InodesUsed (optional)
//   - Mountpoint: from WritableLayer.FsId.Mountpoint if available, empty otherwise.
//     If WritableLayer is nil, an empty FileSystemMetrics struct is returned.
func buildFileSystemMetrics(
	stats *cri.ContainerStats,
) *gen.FileSystemMetrics {
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
