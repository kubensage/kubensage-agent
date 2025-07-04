package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

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
