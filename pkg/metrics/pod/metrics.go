package pod

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func Metrics(pod *cri.PodSandbox, containersMetrics []*gen.ContainerMetrics) (*gen.PodMetrics, error) {
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
