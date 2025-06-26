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
