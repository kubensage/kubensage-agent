package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// listPods fetches a list of pod sandboxes from the runtime service.
// It interacts with the RuntimeServiceClient to retrieve information about all pod sandboxes
// currently managed by the runtime.
func listPods(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.PodSandbox, error) {
	// Call the ListPodSandbox API to fetch pod sandbox details.
	resp, err := runtimeClient.ListPodSandbox(ctx, &runtimeapi.ListPodSandboxRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err.Error())
	}

	// Return the list of pod sandboxes.
	return resp.Items, nil
}

// listPodStats fetches the pod sandbox stats from the runtime service.
// It interacts with the RuntimeServiceClient to retrieve stats for all pod sandboxes
// currently managed by the runtime.
func listPodStats(
	ctx context.Context,
	runtimeClient runtimeapi.RuntimeServiceClient,
) ([]*runtimeapi.PodSandboxStats, error) {
	// Call the ListPodSandboxStats API to fetch stats for all pod sandboxes.
	resp, err := runtimeClient.ListPodSandboxStats(ctx, &runtimeapi.ListPodSandboxStatsRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandbox stats: %v", err.Error())
	}

	// Return the list of pod sandbox stats.
	return resp.Stats, nil
}

// getPodStatsById retrieves the stats for a specific pod sandbox identified by its ID.
// It iterates through the provided list of pod sandbox stats and returns the stats for the pod matching the given ID.
func getPodStatsById(
	podStats []*runtimeapi.PodSandboxStats,
	id string,
) (*runtimeapi.PodSandboxStats, error) {
	// Iterate through the pod sandbox stats to find the one that matches the given ID.
	for _, stats := range podStats {
		// PodSandboxStats.Attributes.Id holds the sandbox ID.
		if stats.Attributes.Id == id {
			return stats, nil
		}
	}

	// Return an error if no stats are found for the specified pod sandbox ID.
	return nil, fmt.Errorf("no pod stats found for sandbox ID %q", id)
}
