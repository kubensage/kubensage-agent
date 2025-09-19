package container

import (
	"time"

	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
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
//   - logger *zap.Logger:
//     Structured logger used for debug tracing during the operation.
//
// Returns:
//
//   - *gen.ContainerMetrics: A protobuf object that consolidates all metrics and metadata
//     related to the container
//   - error: Non-nil if no stats are found for the given container ID. In such cases,
//     the returned metrics object is nil.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func BuildContainerMetrics(
	container *cri.Container,
	stats []*cri.ContainerStats,
	logger *zap.Logger,
) (*gen.ContainerMetrics, error, time.Duration) {
	start := time.Now()

	var retrieveContainerStatsByContainerIdDuration time.Duration
	var buildCpuMetricsDuration time.Duration
	var buildMemoryMetricsDuration time.Duration
	var buildFileSystemMetricsDuration time.Duration
	var buildSwapMetricsDuration time.Duration

	containerStats, err, retrieveContainerStatsByContainerIdDuration := RetrieveContainerStatsByContainerId(stats, container.Id)

	if err != nil {
		return nil, err, time.Since(start)
	}

	cpuMetrics, buildCpuMetricsDuration := buildCpuMetrics(containerStats)
	memoryMetrics, buildMemoryMetricsDuration := buildMemoryMetrics(containerStats)
	fileSystemMetrics, buildFileSystemMetricsDuration := buildFileSystemMetrics(containerStats)
	swapMetrics, buildSwapMetricsDuration := buildSwapMetrics(containerStats)

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

	logger.Debug("container metrics durations",
		zap.String("container_id", container.Id),
		zap.Duration("retrieve_stats_duration", retrieveContainerStatsByContainerIdDuration),
		zap.Duration("build_cpu_duration", buildCpuMetricsDuration),
		zap.Duration("build_memory_duration", buildMemoryMetricsDuration),
		zap.Duration("build_filesystem_duration", buildFileSystemMetricsDuration),
		zap.Duration("build_swap_duration", buildSwapMetricsDuration),
		zap.Duration("total_duration", time.Since(start)),
	)

	return containerMetrics, nil, time.Since(start)
}
