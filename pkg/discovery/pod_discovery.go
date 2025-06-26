package discovery

import (
	"context"
	"fmt"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"time"
)

func ListPods(runtimeClient runtimeapi.RuntimeServiceClient) ([]*runtimeapi.PodSandbox, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := runtimeClient.ListPodSandbox(ctx, &runtimeapi.ListPodSandboxRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err.Error())
	}

	return resp.Items, nil
}

func ListPodStats(runtimeClient runtimeapi.RuntimeServiceClient, podSandboxId string) (*runtimeapi.PodSandboxStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	podSandboxStatsFilter := runtimeapi.PodSandboxStatsFilter{Id: podSandboxId}
	podSandboxStatsRequest := runtimeapi.ListPodSandboxStatsRequest{Filter: &podSandboxStatsFilter}
	resp, err := runtimeClient.ListPodSandboxStats(ctx, &podSandboxStatsRequest)

	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandbox stats: %v", err.Error())
	}

	if len(resp.Stats) == 0 {
		return nil, fmt.Errorf("no pod stats found for pod sandbox id %s", podSandboxId)
	}

	return resp.Stats[0], nil
}
