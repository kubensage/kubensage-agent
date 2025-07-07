package converter

import (
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

func convertContainerMetrics(containers []*metrics.ContainerMetrics) []*pb.ContainerMetrics {
	var result []*pb.ContainerMetrics

	for _, cm := range containers {
		cpu := &pb.CpuMetrics{
			Timestamp:            cm.CpuMetrics.Timestamp,
			UsageNanoCores:       cm.CpuMetrics.UsageNanoCores,
			UsageCoreNanoSeconds: cm.CpuMetrics.UsageCoreNanoSeconds,
		}
		mem := &pb.MemoryMetrics{
			Timestamp:       cm.MemoryMetrics.Timestamp,
			WorkingSetBytes: cm.MemoryMetrics.WorkingSetBytes,
			AvailableBytes:  cm.MemoryMetrics.AvailableBytes,
			UsageBytes:      cm.MemoryMetrics.UsageBytes,
			RssBytes:        cm.MemoryMetrics.RssBytes,
			PageFaults:      cm.MemoryMetrics.PageFaults,
			MajorPageFaults: cm.MemoryMetrics.MajorPageFaults,
		}
		fs := &pb.FileSystemMetrics{
			Timestamp:  cm.FileSystemMetrics.Timestamp,
			Mountpoint: cm.FileSystemMetrics.Mountpoint,
			UsedBytes:  cm.FileSystemMetrics.UsedBytes,
			InodesUsed: cm.FileSystemMetrics.InodesUsed,
		}
		swap := &pb.SwapMetrics{
			Timestamp:      cm.SwapMetrics.Timestamp,
			AvailableBytes: cm.SwapMetrics.SwapAvailableBytes,
			UsageBytes:     cm.SwapMetrics.SwapUsageBytes,
		}

		result = append(result, &pb.ContainerMetrics{
			Id:                cm.Id,
			Name:              cm.Name,
			Image:             cm.Image,
			CreatedAt:         cm.CreatedAt,
			State:             cm.State,
			Attempt:           cm.Attempt,
			CpuMetrics:        cpu,
			MemoryMetrics:     mem,
			FileSystemMetrics: fs,
			SwapMetrics:       swap,
		})
	}

	return result
}
