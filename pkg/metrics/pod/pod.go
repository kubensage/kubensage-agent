package pod

import (
	"context"
	"fmt"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// Pods retrieves a list of all pod sandboxes from the CRI runtime.
// It sends a ListPodSandboxRequest with default filters and returns the response items.
// If the request fails, it returns an error.
func Pods(
	ctx context.Context,
	runtimeClient cri.RuntimeServiceClient,
	getOnlyReady bool,
) ([]*cri.PodSandbox, error) {
	PodSandboxFilter := &cri.PodSandboxFilter{State: &cri.PodSandboxStateValue{State: cri.PodSandboxState_SANDBOX_READY}}

	var resp *cri.ListPodSandboxResponse
	var err error

	if getOnlyReady {
		resp, err = runtimeClient.ListPodSandbox(ctx, &cri.ListPodSandboxRequest{Filter: PodSandboxFilter})
	} else {
		resp, err = runtimeClient.ListPodSandbox(ctx, &cri.ListPodSandboxRequest{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err.Error())
	}
	return resp.Items, nil
}
