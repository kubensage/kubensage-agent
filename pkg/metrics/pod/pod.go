package pod

import (
	"context"
	"fmt"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// ListPods retrieves the list of pod sandboxes from the CRI runtime.
//
// If getOnlyReady is true, the function applies a filter to include only pods
// in the SANDBOX_READY state. If false, it fetches all available pod sandboxes.
//
// Parameters:
//   - ctx context.Context:
//     Standard context for cancellation and timeout propagation.
//   - runtimeClient cri.RuntimeServiceClient:
//     The CRI client used to interact with the container runtime's gRPC API.
//   - getOnlyReady bool:
//     A boolean flag to determine whether to return only ready pods.
//
// Returns:
//   - []*cri.PodSandbox:
//     A slice of pod sandbox objects returned by the runtime.
//   - error:
//     An error if the CRI request fails, or nil on success.
func ListPods(
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
