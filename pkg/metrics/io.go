package metrics

import (
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type IoMetrics struct {
	PsiSome PsiMetrics `json:"psi_some_metrics,omitempty"`
	PsiFull PsiMetrics `json:"psi_full_metrics,omitempty"`
}

func SafeIoMetrics(stats *runtimeapi.ContainerStats) IoMetrics {
	if stats.Swap == nil {
		return IoMetrics{}
	}

	metrics := IoMetrics{
		PsiSome: SafePsiSomeIoMetrics(stats),
		PsiFull: SafePsiFullIoMetrics(stats),
	}

	return metrics
}
