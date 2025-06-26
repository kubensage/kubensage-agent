package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"time"
)

func ListContainers(runtimeClient runtimeapi.RuntimeServiceClient, podSandboxId string) ([]*runtimeapi.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerFilter := runtimeapi.ContainerFilter{PodSandboxId: podSandboxId}
	containersRequest := runtimeapi.ListContainersRequest{Filter: &containerFilter}

	resp, err := runtimeClient.ListContainers(ctx, &containersRequest)

	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err.Error())
	}

	return resp.Containers, nil
}

func ListContainerStats(runtimeClient runtimeapi.RuntimeServiceClient, podSandboxId string, containerId string) ([]*runtimeapi.ContainerStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerStatsFilter := runtimeapi.ContainerStatsFilter{Id: containerId, PodSandboxId: podSandboxId}
	containerStatsRequest := runtimeapi.ListContainerStatsRequest{Filter: &containerStatsFilter}

	stats, err := runtimeClient.ListContainerStats(ctx, &containerStatsRequest)

	if err != nil {
		return nil, fmt.Errorf("failed to list container stats: %v", err.Error())
	}

	return stats.Stats, nil
}
