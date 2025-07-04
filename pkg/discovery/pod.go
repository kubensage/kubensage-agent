package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// getPods retrieves a list of all pod sandboxes from the CRI runtime.
// It sends a ListPodSandboxRequest with default filters and returns the response items.
// If the request fails, it returns an error.
func getPods(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.PodSandbox, error) {
	resp, err := runtimeClient.ListPodSandbox(ctx, &runtimeapi.ListPodSandboxRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err.Error())
	}
	return resp.Items, nil
}

// getPodStats retrieves statistics for all pod sandboxes from the CRI runtime.
// It sends a ListPodSandboxStatsRequest and returns the collected stats slice.
// If the request fails, it returns an error.
func getPodStats(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.PodSandboxStats, error) {
	resp, err := runtimeClient.ListPodSandboxStats(ctx, &runtimeapi.ListPodSandboxStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandbox stats: %v", err.Error())
	}
	return resp.Stats, nil
}

// getPodStatsById searches for the stats of a specific pod sandbox by ID within a list of stats.
// If a match is found, it returns the corresponding PodSandboxStats.
// If not found, it returns an error.
func getPodStatsById(
	podStats []*runtimeapi.PodSandboxStats,
	id string,
) (*runtimeapi.PodSandboxStats, error) {
	for _, stats := range podStats {
		if stats.Attributes.Id == id {
			return stats, nil
		}
	}
	return nil, fmt.Errorf("no pod stats found for sandbox ID %q", id)
}
