package discovery

import (
	"context"
	"fmt"
	m "github.com/kubensage/kubensage-agent/pkg/metrics"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sync"
)

func GetAllMetrics(ctx context.Context, runtimeClient cri.RuntimeServiceClient) (*m.Metrics, []error) {
	var wg sync.WaitGroup

	errChan := make(chan error, 4)

	var pods []*cri.PodSandbox
	var containers []*cri.Container
	var containersStats []*cri.ContainerStats
	var nodeMetrics *m.NodeMetrics

	wg.Add(4)

	go func() {
		defer wg.Done()
		var err error
		nodeMetrics, err = m.SafeNodeMetrics(ctx)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers stats: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		pods, err = getPods(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list pod sandboxes: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		containers, err = getContainers(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		containersStats, err = getContainersStats(ctx, runtimeClient)
		if err != nil {
			errChan <- fmt.Errorf("failed to list containers stats: %v", err)
		}
	}()

	wg.Wait()

	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	var podsMetrics []*m.PodMetrics

	containerMap := make(map[string][]*cri.Container)
	for _, container := range containers {
		containerMap[container.PodSandboxId] = append(containerMap[container.PodSandboxId], container)
	}

	for _, pod := range pods {
		var containersMetrics []*m.ContainerMetrics

		containers := containerMap[pod.Id]

		for _, container := range containers {
			containerStats, err := getContainerStatsByContainerId(containersStats, container.Id)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get container stats for container %s: %v", container.Id, err))
			}

			cpuMetrics := m.SafeCpuMetrics(containerStats)
			memoryMetrics := m.SafeMemoryMetrics(containerStats)
			fileSystemMetrics := m.SafeFileSystemMetrics(containerStats)
			swapMetrics := m.SafeSwapMetrics(containerStats)

			containerMetrics := &m.ContainerMetrics{
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

		podMetric := &m.PodMetrics{
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

	metrics := &m.Metrics{NodeMetrics: nodeMetrics, PodMetrics: podsMetrics}

	return metrics, errs
}
