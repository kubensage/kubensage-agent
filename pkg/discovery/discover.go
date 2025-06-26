package discovery

import (
	"context"
	"fmt"
	"github.com/kubensage/kubensage-agent/pkg/model"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"time"
)

func Discover(ctx context.Context, runtimeClient runtimeapi.RuntimeServiceClient) ([]model.PodInfo, error) {
	// Discover the pods
	podSandboxes, err := listPods(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err)
	}

	var allPodInfo []model.PodInfo

	// Iterate through the pods
	for _, sandbox := range podSandboxes {
		podInfo := model.PodInfo{Timestamp: time.Now().UnixNano(), Pod: sandbox}

		// Retrieve the pod stats
		podStats, err := listPodStats(ctx, runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to list pod stats for sandbox %s: %v", sandbox.Id, err)
		} else {
			podInfo.PodStats = podStats
		}

		// Retrieve the containers of the pod
		containers, err := listContainers(ctx, runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to discover containers for sandbox %s: %v", sandbox.Id, err)
			continue
		}

		var containerInfos []*model.ContainerInfo

		// Retrieve the stats for each container
		for _, container := range containers {
			stats, err := listContainerStats(ctx, runtimeClient, sandbox.Id, container.Id)
			if err != nil {
				log.Printf("Failed to discover stats for container %s: %v", container.Id, err)
				continue
			}

			containerInfo := &model.ContainerInfo{
				Container:      container,
				ContainerStats: stats,
			}

			containerInfos = append(containerInfos, containerInfo)
		}

		podInfo.Containers = containerInfos

		allPodInfo = append(allPodInfo, podInfo)
	}

	return allPodInfo, nil
}
