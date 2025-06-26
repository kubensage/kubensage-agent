package discovery

import (
	"context"
	"fmt"
	"github.com/kubensage/kubensage-agent/pkg/model"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"time"
)

// Discover discovers information about the pods and containers in the runtime environment.
// It fetches pod sandboxes, pod stats, container details, and container stats, then assembles
// this data into a comprehensive list of PodInfo structures, each containing information about
// the pod and its containers, including the stats of each container.
func Discover(ctx context.Context, runtimeClient runtimeapi.RuntimeServiceClient) ([]model.PodInfo, error) {
	var allPodInfo []model.PodInfo

	// List all pods in the runtime environment.
	pods, err := listPods(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err)
	}

	// List the stats for the pods.
	podsStats, err := listPodStats(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod stats: %v", err)
	}

	// List all containers in the runtime environment.
	containers, err := listContainers(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	// List the stats for the containers.
	containersStats, err := listContainersStats(ctx, runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers stats: %v", err)
	}

	// Iterate over each pod and collect associated information.
	for _, pod := range pods {
		// Initialize a PodInfo object to store pod and container details.
		var podInfo model.PodInfo
		podInfo.Timestamp = time.Now().UnixNano() // Set the timestamp for when the discovery occurred.
		podInfo.Pod = pod                         // Set the current pod in the PodInfo object.

		// Retrieve the statistics for the current pod.
		podStats, err := getPodStatsById(podsStats, pod.Id)
		if err != nil {
			log.Println("Failed to get pod stats: ", err)
		} else {
			podInfo.PodStats = podStats // Store the pod stats in the PodInfo object.
		}

		// Get all containers associated with the current pod.
		containers := getContainersByPodID(containers, pod.Id)

		// Initialize a slice to hold container information.
		var containerInfos []*model.ContainerInfo

		// Retrieve the statistics for each container within the pod.
		for _, container := range containers {
			containerStats, err := getContainerStatsByContainerId(containersStats, container.Id)
			if err != nil {
				log.Printf("Failed to discover stats for container %s: %v", container.Id, err)
				continue // Skip this container if stats retrieval fails.
			}

			// Create a ContainerInfo object to store container details and stats.
			containerInfo := &model.ContainerInfo{
				Container:      container,
				ContainerStats: containerStats,
			}

			// Append the container's info to the list of containerInfos.
			containerInfos = append(containerInfos, containerInfo)
		}

		// Associate the collected container information with the current pod.
		podInfo.Containers = containerInfos

		// Append the completed PodInfo to the list of allPodInfo.
		allPodInfo = append(allPodInfo, podInfo)
	}

	// Return the final list of all pod and container information.
	return allPodInfo, nil
}
