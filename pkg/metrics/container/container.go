package container

import (
	"context"
	"fmt"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// Containers retrieves the list of running containers from the CRI runtime.
// It sends a ListContainersRequest with default filters and returns the result.
// If the request fails, it returns an error.
func Containers(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
) ([]*cri.Container, error) {
	resp, err := runtimeClient.ListContainers(ctx, &cri.ListContainersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error())
	}
	return resp.Containers, nil
}

// ContainersStats fetches statistics for all containers from the CRI runtime.
// It returns a slice of ContainerStats. If no stats are found or the request fails,
// it returns an appropriate error.
func ContainersStats(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
) ([]*cri.ContainerStats, error) {
	stats, err := runtimeClient.ListContainerStats(ctx, &cri.ListContainerStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list container stats: %v", err.Error())
	}

	if len(stats.Stats) == 0 {
		return nil, fmt.Errorf("no container stats found")
	}

	return stats.Stats, nil
}

// ContainerStatsByContainerId searches for and returns the ContainerStats
// matching the given container ID from the provided list.
// If no match is found, it returns an error.
func ContainerStatsByContainerId(
	containersStats []*cri.ContainerStats,
	containerId string,
) (*cri.ContainerStats, error) {
	for _, s := range containersStats {
		if s.Attributes.Id == containerId {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no container stats found for container %q", containerId)
}
