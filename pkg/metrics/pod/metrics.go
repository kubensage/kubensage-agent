package pod

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// BuildPodMetrics constructs a PodMetrics message from the given CRI pod sandbox
// and a list of container metrics associated with that pod.
//
// The resulting PodMetrics contains metadata identifying the pod (ID, name, namespace, UID),
// lifecycle attributes (creation time, state, attempt), and aggregated container-level metrics.
//
// Parameters:
//
//   - pod *cri.PodSandbox:
//     The pod sandbox structure returned by the CRI runtime. This contains metadata such as
//     name, namespace, UID, and lifecycle status.
//
//   - containersMetrics []*gen.ContainerMetrics:
//     A slice of ContainerMetrics representing the metrics collected from the pod's containers.
//
// Returns:
//
//   - *gen.PodMetrics:
//     A fully populated PodMetrics protobuf message.
//
//   - error:
//     Always returns nil in the current implementation, but the signature allows future extension
//     to support validation or error conditions.
func BuildPodMetrics(
	pod *cri.PodSandbox,
	containersMetrics []*gen.ContainerMetrics,
) (*gen.PodMetrics, error) {
	return &gen.PodMetrics{
		Id:               pod.Id,
		Uid:              pod.Metadata.Uid,
		Name:             pod.Metadata.Name,
		Namespace:        pod.Metadata.Namespace,
		CreatedAt:        pod.CreatedAt,
		State:            pod.State.String(),
		Attempt:          pod.Metadata.Attempt,
		ContainerMetrics: containersMetrics,
	}, nil
}
