package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// listContainers fetches a list of containers from the runtime service.
// It interacts with the RuntimeServiceClient to retrieve information about all the containers
// currently managed by the runtime.
func listContainers(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.Container, error) {
	// Call the ListContainers API to fetch container details.
	resp, err := runtimeClient.ListContainers(ctx, &runtimeapi.ListContainersRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error())
	}

	// Return the list of containers.
	return resp.Containers, nil
}

// getContainersByPodID filters and returns containers that are associated with the given podSandboxID.
// It iterates over the list of all containers and filters those whose PodSandboxId matches the provided ID.
func getContainersByPodID(
	containers []*runtimeapi.Container,
	podSandboxID string,
) []*runtimeapi.Container {
	var filtered []*runtimeapi.Container
	// Iterate through containers and select those associated with the given PodSandboxId.
	for _, ctr := range containers {
		// The PodSandboxId is stored on each container's attributes.
		if ctr.PodSandboxId == podSandboxID {
			filtered = append(filtered, ctr)
		}
	}

	// Return the filtered list of containers.
	return filtered
}

// listContainersStats fetches the container stats from the runtime service.
// It interacts with the RuntimeServiceClient to retrieve the stats of all containers currently managed by the runtime.
func listContainersStats(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.ContainerStats, error) {
	// Call the ListContainerStats API to fetch the stats for all containers.
	stats, err := runtimeClient.ListContainerStats(ctx, &runtimeapi.ListContainerStatsRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list container stats: %v", err.Error())
	}

	// If no stats are found, return an error.
	if len(stats.Stats) == 0 {
		return nil, fmt.Errorf("no container stats found")
	}

	// Return the list of container stats.
	return stats.Stats, nil
}

// getContainerStatsByContainerId retrieves the stats for a specific container identified by its containerId.
// It iterates through the provided list of container stats and returns the stats for the container matching the ID.
func getContainerStatsByContainerId(
	containersStats []*runtimeapi.ContainerStats,
	containerId string,
) (*runtimeapi.ContainerStats, error) {
	// Iterate through the container stats to find the one that matches the given containerId.
	for _, s := range containersStats {
		if s.Attributes.Id == containerId {
			return s, nil
		}
	}

	// Return an error if no stats are found for the specified containerId.
	return nil, fmt.Errorf("no container stats found for container %q", containerId)
}
