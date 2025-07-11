package metrics

import (
	"context"
	"fmt"
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sync"
	"time"
)

// GetMetrics collects both node-level and pod-level metrics from the container runtime API (CRI).
// The collection process is parallelized across four sources:
// - Node metrics (CPU, memory, PSI, etc.)
// - Pod sandboxes
// - Containers
// - Container stats
//
// The function returns a populated *Metrics object and a slice of errors that occurred during collection.
// Partial failures (e.g., missing stats for a container) do not block the overall process.
func GetMetrics(ctx context.Context, runtimeClient cri.RuntimeServiceClient, logger *zap.Logger) (*proto.Metrics, []error) {
	var wg sync.WaitGroup

	// Error channel for concurrent metric collection
	errChan := make(chan error, 4)

	var pods []*cri.PodSandbox
	var containers []*cri.Container
	var containersStats []*cri.ContainerStats
	var nodeMetrics *proto.NodeMetrics // NodeMetrics

	wg.Add(4)

	// Collect node-level metrics concurrently
	go func() {
		defer wg.Done()
		var err error
		nodeMetrics, err = getNodeMetrics(ctx, 1*time.Second, logger)
		if err != nil {
			errChan <- fmt.Errorf("failed to collect node metrics: %v", err)
		}
	}()

	// List all pod sandboxes
	go func() {
		defer wg.Done()
		var err error
		pods, err = getPods(ctx, runtimeClient, true)
		if err != nil {
			errChan <- fmt.Errorf("failed to list pod sandboxes: %v", err)
		}
	}()

	// List all containers
	go func() {
		defer wg.Done()
		var err error
		containers, err = getContainers(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers: %v", err)
		}
	}()

	// Get stats for all containers
	go func() {
		defer wg.Done()
		var err error
		containersStats, err = getContainersStats(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list container stats: %v", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Collect all async errors into a slice
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	var podsMetrics []*proto.PodMetrics

	// Create a mapping from pod ID to its associated containers
	containerMap := make(map[string][]*cri.Container)
	for _, container := range containers {
		containerMap[container.PodSandboxId] = append(containerMap[container.PodSandboxId], container)
	}

	// Build pod metrics from pod and container data
	for _, pod := range pods {
		var containersMetrics []*proto.ContainerMetrics
		containers := containerMap[pod.Id]

		for _, container := range containers {
			// Match stats for each container
			containerStats, err := getContainerStatsByContainerId(containersStats, container.Id)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get container stats for container %s: %v", container.Id, err))
			}

			// Extract metrics from stats object (safe access wrappers)
			cpuMetrics := getCpuMetrics(containerStats)
			memoryMetrics := getMemoryMetrics(containerStats)
			fileSystemMetrics := getFileSystemMetrics(containerStats)
			swapMetrics := getSwapMetrics(containerStats)

			// Build container metric object
			containerMetrics := &proto.ContainerMetrics{
				Id:                container.Id,
				Name:              container.Metadata.Name,
				Image:             container.Image.Image,
				CreatedAt:         container.CreatedAt,
				State:             container.State.String(),
				Attempt:           container.Metadata.Attempt,
				CpuMetrics:        cpuMetrics,
				MemoryMetrics:     memoryMetrics,
				FileSystemMetrics: fileSystemMetrics,
				SwapMetrics:       swapMetrics,
			}

			containersMetrics = append(containersMetrics, containerMetrics)
		}

		// Build pod-level metric from collected container metrics
		podMetric := &proto.PodMetrics{
			Id:               pod.Id,
			Uid:              pod.Metadata.Uid,
			Name:             pod.Metadata.Name,
			Namespace:        pod.Metadata.Namespace,
			CreatedAt:        pod.CreatedAt,
			State:            pod.State.String(),
			Attempt:          pod.Metadata.Attempt,
			ContainerMetrics: containersMetrics,
		}

		podsMetrics = append(podsMetrics, podMetric)
	}

	// Final assembled metrics object
	metrics := &proto.Metrics{
		NodeMetrics: nodeMetrics,
		PodMetrics:  podsMetrics,
	}

	return metrics, errs
}
