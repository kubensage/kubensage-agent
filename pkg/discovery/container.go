package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// getContainers retrieves the list of running containers from the CRI runtime.
// It sends a ListContainersRequest with default filters and returns the result.
// If the request fails, it returns an error.
func getContainers(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.Container, error) {
	resp, err := runtimeClient.ListContainers(ctx, &runtimeapi.ListContainersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error())
	}
	return resp.Containers, nil
}

// getContainersStats fetches statistics for all containers from the CRI runtime.
// It returns a slice of ContainerStats. If no stats are found or the request fails,
// it returns an appropriate error.
func getContainersStats(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.ContainerStats, error) {
	stats, err := runtimeClient.ListContainerStats(ctx, &runtimeapi.ListContainerStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list container stats: %v", err.Error())
	}

	if len(stats.Stats) == 0 {
		return nil, fmt.Errorf("no container stats found")
	}

	return stats.Stats, nil
}

// getContainerStatsByContainerId searches for and returns the ContainerStats
// matching the given container ID from the provided list.
// If no match is found, it returns an error.
func getContainerStatsByContainerId(
	containersStats []*runtimeapi.ContainerStats,
	containerId string,
) (*runtimeapi.ContainerStats, error) {
	for _, s := range containersStats {
		if s.Attributes.Id == containerId {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no container stats found for container %q", containerId)
}
