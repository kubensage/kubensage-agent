package metrics

import (
	"context"
	"fmt"
	"gitlab.com/kubensage/kubensage-agent/pkg/metrics/container"
	"gitlab.com/kubensage/kubensage-agent/pkg/metrics/node"
	"gitlab.com/kubensage/kubensage-agent/pkg/metrics/pod"
	"gitlab.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sync"
	"time"
)

// Metrics collects both node-level and pod-level metrics from the container runtime API (CRI).
// The collection process is parallelized across four sources:
// - Node metrics (CPU, memory, PSI, etc.)
// - Pod sandboxes
// - Containers
// - Container stats
//
// The function returns a populated *Metrics object and a slice of errors that occurred during collection.
// Partial failures (e.g., missing stats for a container) do not block the overall process.
func Metrics(ctx context.Context, runtimeClient cri.RuntimeServiceClient, logger *zap.Logger) (*gen.Metrics, []error) {
	var wg sync.WaitGroup

	// Error channel for concurrent metric collection
	errChan := make(chan error, 4)

	var pods []*cri.PodSandbox
	var containers []*cri.Container
	var containersStats []*cri.ContainerStats
	var nodeMetrics *gen.NodeMetrics // NodeMetrics

	wg.Add(4)

	// Collect node-level metrics concurrently
	go func() {
		defer wg.Done()
		var err error
		nodeMetrics, err = node.Metrics(ctx, 1*time.Second, logger)
		if err != nil {
			errChan <- fmt.Errorf("failed to collect node metrics: %v", err)
		}
	}()

	// List all _pod sandboxes
	go func() {
		defer wg.Done()
		var err error
		pods, err = pod.Pods(ctx, runtimeClient, true)
		if err != nil {
			errChan <- fmt.Errorf("failed to list _pod sandboxes: %v", err)
		}
	}()

	// List all containers
	go func() {
		defer wg.Done()
		var err error
		containers, err = container.Containers(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers: %v", err)
		}
	}()

	// Get stats for all containers
	go func() {
		defer wg.Done()
		var err error
		containersStats, err = container.ContainersStats(ctx, runtimeClient)
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

	var podsMetrics []*gen.PodMetrics

	// Create a mapping from _pod ID to its associated containers
	containerMap := make(map[string][]*cri.Container)
	for _, _container := range containers {
		containerMap[_container.PodSandboxId] = append(containerMap[_container.PodSandboxId], _container)
	}

	// Build _pod metrics from _pod and container data
	for _, _pod := range pods {
		var containersMetrics []*gen.ContainerMetrics
		containers := containerMap[_pod.Id]

		for _, _container := range containers {
			metrics, err := container.Metrics(_container, containersStats)

			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get _container stats for _container %s: %v", _container.Id, err))
				continue
			}

			containersMetrics = append(containersMetrics, metrics)
		}

		podMetric, _ := pod.Metrics(_pod, containersMetrics)
		podsMetrics = append(podsMetrics, podMetric)
	}

	// Final assembled metrics object
	metrics := &gen.Metrics{
		NodeMetrics: nodeMetrics,
		PodMetrics:  podsMetrics,
	}

	return metrics, errs
}
