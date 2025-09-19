package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubensage/go-common/datastructure"
	gogo "github.com/kubensage/go-common/go"
	"github.com/kubensage/kubensage-agent/pkg/cli"
	"github.com/kubensage/kubensage-agent/pkg/metrics/container"
	"github.com/kubensage/kubensage-agent/pkg/metrics/node"
	"github.com/kubensage/kubensage-agent/pkg/metrics/pod"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
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
	start := time.Now()
	logger.Info("collect start", zap.Int("topN", agentCfg.TopN))

	metricsData, errs := collect(ctx, runtimeClient, logger, agentCfg.TopN)

	// Se errori, log ERROR + breve riepilogo INFO
	if errs != nil && len(errs) > 0 {
		errStrs := make([]string, 0, len(errs))
		for _, e := range errs {
			errStrs = append(errStrs, e.Error())
		}
		logger.Error("metric collection errors", zap.Strings("errors", errStrs))
	}

	// Evita panic se metricsData è nil
	if metricsData == nil {
		logger.Info("collect finished (no metrics)", zap.Duration("duration", time.Since(start)))
		return errs
	}

	podsCount := 0
	containersCount := 0
	if metricsData.PodMetrics != nil {
		podsCount = len(metricsData.PodMetrics)
		for _, pm := range metricsData.PodMetrics {
			if pm != nil && pm.ContainerMetrics != nil {
				containersCount += len(pm.ContainerMetrics)
			}
		}
	}

	buffer.Add(metricsData)

	logger.Info("collect enqueued",
		zap.Int64("timestamp", metricsData.Timestamp),
		zap.Int("pods_count", podsCount),
		zap.Int("containers_count", containersCount),
		zap.Int("buffer_len", buffer.Len()),
		zap.Duration("duration", time.Since(start)),
	)

	return errs
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
	start := time.Now()
	timestamp := time.Now().Unix()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	addErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		errs = append(errs, e)
		mu.Unlock()
	}

	// Durations (macro)
	var buildNodeMetricsDuration time.Duration
	var listPodsDuration time.Duration
	var listContainerDuration time.Duration
	var listContainersStatsDuration time.Duration

	// Durations (post-processing)
	var buildContainerMetricsTotalDuration time.Duration
	var buildPodMetricsTotalDuration time.Duration

	// Per-debug: container più lento
	var slowestContainerID string
	var slowestContainerDuration time.Duration

	// ===== parallel fetch =====
	var nodeMetrics *gen.NodeMetrics
	gogo.SafeGo(&wg, func() {
		var subErrs []error
		var d time.Duration
		nodeMetrics, subErrs, d = node.BuildNodeMetrics(ctx, 0*time.Second, logger, topN)
		buildNodeMetricsDuration = d
		if subErrs != nil {
			for _, e := range subErrs {
				addErr(e)
			}
		}
	})

	var pods []*cri.PodSandbox
	gogo.SafeGo(&wg, func() {
		var err error
		var d time.Duration
		pods, err, d = pod.ListPods(ctx, runtimeClient, true)
		listPodsDuration = d
		addErr(err)
	})

	var containers []*cri.Container
	gogo.SafeGo(&wg, func() {
		var err error
		var d time.Duration
		containers, err, d = container.ListContainers(ctx, runtimeClient)
		listContainerDuration = d
		addErr(err)
	})

	var containersStats []*cri.ContainerStats
	gogo.SafeGo(&wg, func() {
		var err error
		var d time.Duration
		containersStats, err, d = container.ListContainersStats(ctx, runtimeClient)
		listContainersStatsDuration = d
		addErr(err)
	})

	wg.Wait()

	// ===== correlate & build =====
	var podsMetrics []*gen.PodMetrics

	containerMap := make(map[string][]*cri.Container)
	for _, c := range containers {
		containerMap[c.PodSandboxId] = append(containerMap[c.PodSandboxId], c)
	}

	for _, p := range pods {
		var containersMetrics []*gen.ContainerMetrics
		cs := containerMap[p.Id]

		for _, c := range cs {
			metrics, err, d := container.BuildContainerMetrics(c, containersStats, logger)
			if err != nil {
				addErr(fmt.Errorf("failed to get container stats for container %s: %v", c.Id, err))
				continue
			}
			containersMetrics = append(containersMetrics, metrics)

			buildContainerMetricsTotalDuration += d
			if d > slowestContainerDuration {
				slowestContainerDuration = d
				slowestContainerID = c.Id
			}
		}

		podStart := time.Now()
		podMetric, _ := pod.BuildPodMetrics(p, containersMetrics)
		buildPodMetricsTotalDuration += time.Since(podStart)

		podsMetrics = append(podsMetrics, podMetric)
	}

	totalDuration := time.Since(start)

	logger.Debug("collect summary",
		zap.Int64("timestamp", timestamp),
		zap.Int("pods_count", len(pods)),
		zap.Int("containers_count", len(containers)),
		zap.Int("containers_stats_count", len(containersStats)),

		zap.Duration("build_node_metrics", buildNodeMetricsDuration),
		zap.Duration("list_pods", listPodsDuration),
		zap.Duration("list_containers", listContainerDuration),
		zap.Duration("list_containers_stats", listContainersStatsDuration),

		zap.Duration("build_container_metrics_total", buildContainerMetricsTotalDuration),
		zap.Duration("build_pod_metrics_total", buildPodMetricsTotalDuration),

		zap.String("slowest_container_id", slowestContainerID),
		zap.Duration("slowest_container_duration", slowestContainerDuration),

		zap.Duration("total_duration", totalDuration),
	)

	if len(errs) > 0 {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		logger.Info("collect completed with errors",
			zap.Int("errors_count", len(errs)),
			zap.Strings("errors", msgs),
		)
	} else {
		logger.Info("collect completed successfully")
	}

	metrics := &gen.Metrics{
		Timestamp:   timestamp,
		NodeMetrics: nodeMetrics,
		PodMetrics:  podsMetrics,
	}

	return metrics, errs
}
