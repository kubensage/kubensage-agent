package converter

import (
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

func convertPodMetrics(pods []*metrics.PodMetrics) []*pb.PodMetrics {
	var result []*pb.PodMetrics

	for _, pm := range pods {
		p := &pb.PodMetrics{
			Id:               pm.Id,
			Uid:              pm.Uid,
			Name:             pm.Name,
			Namespace:        pm.Namespace,
			CreatedAt:        pm.CreatedAt,
			State:            pm.State,
			Attempt:          pm.Attempt,
			ContainerMetrics: convertContainerMetrics(pm.ContainerMetrics),
		}
		result = append(result, p)
	}

	return result
}
