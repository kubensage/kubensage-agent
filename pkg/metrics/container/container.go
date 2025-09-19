package container

import (
	"context"
	"fmt"
	"time"

	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// ListContainers retrieves all containers currently known to the CRI runtime.
//
// This function performs an unfiltered ListContainers RPC call to the container runtime,
// meaning it returns all containers regardless of their state (e.g., running, exited, etc.).
//
// Parameters:
//   - ctx: context.Context - the context used to control cancellation and deadlines for the RPC call.
//   - runtimeClient: cri.RuntimeServiceClient - the CRI runtime client used to issue the ListContainers request.
//
// Returns:
//   - []*cri.Container: a slice of container descriptors, each containing metadata and runtime state.
//   - error: if the RPC call fails or returns an unexpected response.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func ListContainers(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
) ([]*cri.Container, error, time.Duration) {
	start := time.Now()

	resp, err := runtimeClient.ListContainers(ctx, &cri.ListContainersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error()), time.Since(start)
	}

	return resp.Containers, nil, time.Since(start)
}

// ListContainersStats retrieves runtime statistics for all containers managed by the CRI runtime.
//
// This function invokes the ListContainerStats RPC without any filters, collecting metrics
// such as CPU, memory, I/O, and filesystem usage. If the response is empty, it returns an explicit error.
//
// Parameters:
//   - ctx: context.Context - used to control cancellation and timeouts for the RPC call.
//   - runtimeClient: cri.RuntimeServiceClient - the CRI client used to issue the stats request.
//
// Returns:
//   - []*cri.ContainerStats: a slice of container statistics, each representing a container's resource usage.
//   - error: if the RPC call fails or no statistics are returned.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func ListContainersStats(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
) ([]*cri.ContainerStats, error, time.Duration) {
	start := time.Now()

	stats, err := runtimeClient.ListContainerStats(ctx, &cri.ListContainerStatsRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list container stats: %v", err.Error()), time.Since(start)
	}

	if len(stats.Stats) == 0 {
		return nil, fmt.Errorf("no container stats found"), time.Since(start)
	}

	return stats.Stats, nil, time.Since(start)
}

// RetrieveContainerStatsByContainerId searches for statistics of a specific container by its ID.
//
// This function iterates over a pre-fetched list of container statistics and returns the matching
// entry for the given container ID. It is useful for mapping container metadata to its corresponding metrics.
//
// Parameters:
//   - containersStats: []*cri.ContainerStats - a slice of container stats to search through.
//   - containerId: string - the ID of the container whose stats are being queried.
//
// Returns:
//   - *cri.ContainerStats: the matching stats entry, if found.
//   - error: if no matching container stats are found for the given ID.
//   - time.Duration: the total time taken to complete the function, useful for performance monitoring.
func RetrieveContainerStatsByContainerId(
	containersStats []*cri.ContainerStats,
	containerId string,
) (*cri.ContainerStats, error, time.Duration) {
	start := time.Now()

	for _, s := range containersStats {
		if s.Attributes.Id == containerId {
			return s, nil, time.Since(start)
		}
	}
	return nil, fmt.Errorf("no container stats found for container %q", containerId), time.Since(start)
}
