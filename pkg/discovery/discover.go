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
	var allPodInfo []model.PodInfo

	// Discover the pods
	pods, err := listPods(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err)
	}

	podsStats, err := listPodStats(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod stats: %v", err)
	}

	containers, err := listContainers(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	containersStats, err := listContainersStats(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list containersStats: %v", err)
	}

	// Iterate through the pods
	for _, pod := range pods {
		// PodInfo initialization
		var podInfo model.PodInfo
		podInfo.Timestamp = time.Now().UnixNano()
		podInfo.Pod = pod

		// Retrieve the pod stats
		podStats, err := getPodStatsById(podsStats, pod.Id)
		if err != nil {
			log.Println("Failed to get pod stats: ", err)
		} else {
			podInfo.PodStats = podStats
		}

		containers := getContainersByPodID(containers, pod.Id)

		var containerInfos []*model.ContainerInfo

		// Retrieve the stats for each container
		for _, container := range containers {
			containerStats, err := getContainerStatsByContainerId(containersStats, container.Id)
			if err != nil {
				log.Printf("Failed to discover stats for container %s: %v", container.Id, err)
				continue
			}

			containerInfo := &model.ContainerInfo{
				Container:      container,
				ContainerStats: containerStats,
			}

			containerInfos = append(containerInfos, containerInfo)
		}

		podInfo.Containers = containerInfos

		allPodInfo = append(allPodInfo, podInfo)
	}

	return allPodInfo, nil
}
