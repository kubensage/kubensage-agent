package metrics

import (
	"context"
	"fmt"
	"github.com/kubensage/go-common/datastructure"
	gogo "github.com/kubensage/go-common/go"
	"github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/pkg/metrics/container"
	"github.com/kubensage/kubensage-agent/pkg/metrics/node"
	"github.com/kubensage/kubensage-agent/pkg/metrics/pod"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sync"
	"time"
)

// CollectOnce performs a single metrics collection cycle.
//
// It retrieves metrics from the node, containers, and pods by calling the internal `collect`
// function. The gathered metrics are then pushed into the provided ring buffer for later transmission.
//
// This function is typically invoked periodically by the main loop.
//
// Parameters:
//   - ctx context.Context:
//     Context for managing timeouts or cancellation of the metric collection process.
//   - runtimeClient cri.RuntimeServiceClient:
//     CRI client used to query the container runtime for pods, containers, and stats.
//   - buffer *datastructure.RingBuffer[*gen.Metrics]:
//     A ring buffer where the collected *gen.Metrics data is stored.
//   - agentCfg *cli.AgentConfig:
//     Agent configuration, including the number of top memory-consuming processes to collect.
//   - logger *zap.Logger:
//     Logger instance for debug and error output.
//
// Returns:
//   - []error:
//     A slice of errors encountered during metric collection.
//     Returns nil if collection succeeded without any issues.
func CollectOnce(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
	buffer *datastructure.RingBuffer[*gen.Metrics],
	agentCfg *cli.AgentConfig,
	logger *zap.Logger,
) []error {
	logger.Debug("Collecting metrics")
	metricsData, errs := collect(ctx, runtimeClient, logger, agentCfg.TopN)
	if errs != nil {
		var errStrs []string
		for _, e := range errs {
			errStrs = append(errStrs, e.Error())
		}
		logger.Error("Metric collection errors", zap.Strings("errors", errStrs))
		return errs
	}
	buffer.Add(metricsData)
	logger.Debug("metrics added to buffer", zap.Int("buffer_len", buffer.Len()))

	return nil
}

// collect gathers both node-level and container-level metrics in parallel,
// using goroutines for efficient concurrent execution. It constructs a
// unified *gen.Metrics message containing resource usage data from the node,
// pods, and their containers.
//
// This function performs the following in parallel:
//   - Collects node metrics (CPU, memory, disk, network, etc.)
//   - Lists running pods
//   - Lists running containers
//   - Fetches container stats
//
// Then it aggregates container metrics per pod, builds corresponding
// *gen.PodMetrics, and wraps everything into a *gen.Metrics structure.
//
// Parameters:
//   - ctx context.Context:
//     Context for managing cancellation and timeouts.
//   - runtimeClient cri.RuntimeServiceClient:
//     CRI client interface for interacting with the container runtime.
//   - logger *zap.Logger:
//     Logger instance used for debugging and error reporting.
//   - topN int:
//     Number of top memory-consuming processes to include in node metrics.
//
// Returns:
//   - *gen.Metrics:
//     A metrics object containing node-wide and per-pod/container metrics.
//   - []error:
//     A list of errors encountered during metric collection. May be empty.
func collect(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
	logger *zap.Logger,
	topN int,
) (*gen.Metrics, []error) {
	timestamp := time.Now().Unix()

	var wg sync.WaitGroup
	var errs []error

	var nodeMetrics *gen.NodeMetrics // NodeMetrics
	gogo.SafeGo(&wg, func() {
		var errs []error
		nodeMetrics, errs = node.BuildNodeMetrics(ctx, 0*time.Second, logger, topN)
		if errs != nil {
			for _, e := range errs {
				errs = append(errs, e)
			}
		}
	})

	var pods []*cri.PodSandbox
	gogo.SafeGo(&wg, func() {
		var err error
		pods, err = pod.ListPods(ctx, runtimeClient, true)
		if err != nil {
			errs = append(errs, err)
		}
	})

	var containers []*cri.Container
	gogo.SafeGo(&wg, func() {
		var err error
		containers, err = container.ListContainers(ctx, runtimeClient)
		if err != nil {
			errs = append(errs, err)
		}
	})

	var containersStats []*cri.ContainerStats
	gogo.SafeGo(&wg, func() {
		var err error
		containersStats, err = container.ListContainersStats(ctx, runtimeClient)
		if err != nil {
			errs = append(errs, err)
		}
	})

	wg.Wait()

	var podsMetrics []*gen.PodMetrics

	containerMap := make(map[string][]*cri.Container)
	for _, _container := range containers {
		containerMap[_container.PodSandboxId] = append(containerMap[_container.PodSandboxId], _container)
	}

	for _, _pod := range pods {
		var containersMetrics []*gen.ContainerMetrics
		containers := containerMap[_pod.Id]

		for _, _container := range containers {
			metrics, err := container.BuildContainerMetrics(_container, containersStats)

			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get _container stats for _container %s: %v", _container.Id, err))
				continue
			}

			containersMetrics = append(containersMetrics, metrics)
		}

		podMetric, _ := pod.BuildPodMetrics(_pod, containersMetrics)
		podsMetrics = append(podsMetrics, podMetric)
	}

	metrics := &gen.Metrics{
		Timestamp:   timestamp,
		NodeMetrics: nodeMetrics,
		PodMetrics:  podsMetrics,
	}

	return metrics, errs
}
