package container

import (
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func Metrics(container *cri.Container, stats []*cri.ContainerStats) (*gen.ContainerMetrics, error) {
	// Match stats for each container
	containerStats, err := ContainerStatsByContainerId(stats, container.Id)
	if err != nil {
		return nil, err
	}

	// Extract metrics from stats object (safe access wrappers)
	cpuMetrics := cpuMetrics(containerStats)
	memoryMetrics := memoryMetrics(containerStats)
	fileSystemMetrics := fileSystemMetrics(containerStats)
	swapMetrics := swapMetrics(containerStats)

	// Build container metric object
	containerMetrics := &gen.ContainerMetrics{
		Id:                container.Id,
		Name:              container.Metadata.Name,
		Image:             container.Image.Image,
		CreatedAt:         container.CreatedAt,
		State:             container.State.String(),
		Attempt:           container.Metadata.Attempt,
		CpuMetrics:        cpuMetrics,
		MemoryMetrics:     memoryMetrics,
		FileSystemMetrics: fileSystemMetrics,
		SwapMetrics:       swapMetrics,
	}

	return containerMetrics, nil
}
