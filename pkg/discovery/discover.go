package discovery

import (
	"context"
	"fmt"
	"github.com/kubensage/kubensage-agent/pkg/model"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"sync"
	"time"
)

// Discover discovers information about the pods and containers in the runtime environment.
func Discover(ctx context.Context, runtimeClient runtimeapi.RuntimeServiceClient) ([]model.PodInfo, error) {
	var allPodInfo []model.PodInfo

	// List all the necessary resources in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// Fetch pods, pod stats, containers, and container stats in parallel
	var pods []*runtimeapi.PodSandbox
	var podsStats []*runtimeapi.PodSandboxStats
	var containers []*runtimeapi.Container
	var containersStats []*runtimeapi.ContainerStats

	wg.Add(4)

	// Fetch pods
	go func() {
		defer wg.Done()
		var err error
		pods, err = listPods(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list pod sandboxes: %v", err)
		}
	}()

	// Fetch pod stats
	go func() {
		defer wg.Done()
		var err error
		podsStats, err = listPodStats(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list pod stats: %v", err)
		}
	}()

	// Fetch containers
	go func() {
		defer wg.Done()
		var err error
		containers, err = listContainers(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers: %v", err)
		}
	}()

	// Fetch container stats
	go func() {
		defer wg.Done()
		var err error
		containersStats, err = listContainersStats(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers stats: %v", err)
		}
	}()

	// Wait for all fetches to complete
	wg.Wait()

	// If any error occurred during parallel fetching, return it
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	// Map containers by Pod ID for faster lookup
	containerMap := make(map[string][]*runtimeapi.Container)
	for _, container := range containers {
		containerMap[container.PodSandboxId] = append(containerMap[container.PodSandboxId], container)
	}

	// Iterate over each pod and collect associated information
	for _, pod := range pods {
		// Initialize a PodInfo object to store pod and container details
		var podInfo model.PodInfo
		podInfo.Timestamp = time.Now().UnixNano() // Set the timestamp for when the discovery occurred
		podInfo.Pod = pod                         // Set the current pod in the PodInfo object

		/* TODO understand what we want to do with pod having state NOT_READY
		if pod.State == runtimeapi.PodSandboxState_SANDBOX_NOTREADY {
			log.Printf("Found pod not ready %s, skipping it...", pod.Id)
			continue
		}*/

		// Retrieve the statistics for the current pod
		podStats, err := getPodStatsById(podsStats, pod.Id)
		if err != nil {
			log.Printf("Failed to get pod stats: %v, pod state %v", err, pod.State)
		} else {
			podInfo.PodStats = podStats // Store the pod stats in the PodInfo object
		}

		// Get all containers associated with the current pod (using pre-built map for faster lookup)
		containers := containerMap[pod.Id]

		// Initialize a slice to hold container information
		var containerInfos []*model.ContainerInfo

		// Retrieve the statistics for each container within the pod
		for _, container := range containers {
			containerStats, err := getContainerStatsByContainerId(containersStats, container.Id)
			if err != nil {
				log.Printf("Failed to discover stats for container %s: %v", container.Id, err)
				continue // Skip this container if stats retrieval fails
			}

			// Create a ContainerInfo object to store container details and stats
			containerInfo := &model.ContainerInfo{
				Container:      container,
				ContainerStats: containerStats,
			}

			// Append the container's info to the list of containerInfos
			containerInfos = append(containerInfos, containerInfo)
		}

		// Associate the collected container information with the current pod
		podInfo.Containers = containerInfos

		// Append the completed PodInfo to the list of allPodInfo
		allPodInfo = append(allPodInfo, podInfo)
	}

	// Return the final list of all pod and container information
	return allPodInfo, nil
}
