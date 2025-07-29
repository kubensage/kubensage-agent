package container

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// BuildContainerMetrics constructs a full *gen.ContainerMetrics message by combining
// metadata from a CRI Container object and resource statistics from ContainerStats.
//
// The function looks up the container's stats by its ID from the provided slice of
// *cri.ContainerStats. If a matching stats object is found, it extracts and builds
// metrics for CPU, memory, filesystem, and swap using internal helpers.
//
// Parameters:
//   - container: *cri.Container
//     The container definition as returned by the CRI runtime, containing metadata
//     such as name, image, creation timestamp, and state.
//   - stats: []*cri.ContainerStats
//     A slice of all available container statistics from the runtime. The function
//     searches for the matching entry using container.Id.
//
// Returns:
//
//   - *gen.ContainerMetrics: A protobuf object that consolidates all metrics and metadata
//     related to the container, including:
//
//   - ID, name, image, created time, state, attempt number
//
//   - CpuMetrics, MemoryMetrics, FileSystemMetrics, SwapMetrics
//
//   - error: Non-nil if no stats are found for the given container ID. In such cases,
//     the returned metrics object is nil.
func BuildContainerMetrics(
	container *cri.Container,
	stats []*cri.ContainerStats,
) (*gen.ContainerMetrics, error) {
	containerStats, err := RetrieveContainerStatsByContainerId(stats, container.Id)
	if err != nil {
		return nil, err
	}

	cpuMetrics := buildCpuMetrics(containerStats)
	memoryMetrics := buildMemoryMetrics(containerStats)
	fileSystemMetrics := buildFileSystemMetrics(containerStats)
	swapMetrics := buildSwapMetrics(containerStats)

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
