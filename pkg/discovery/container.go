package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func listContainers(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.Container, error) {
	resp, err := runtimeClient.ListContainers(ctx, &runtimeapi.ListContainersRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error())
	}

	return resp.Containers, nil
}

func getContainersByPodID(
	containers []*runtimeapi.Container,
	podSandboxID string,
) []*runtimeapi.Container {
	var filtered []*runtimeapi.Container
	for _, ctr := range containers {
		// The PodSandboxId is stored on each Container’s attributes
		if ctr.PodSandboxId == podSandboxID {
			filtered = append(filtered, ctr)
		}
	}

	return filtered
}

func listContainersStats(
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
