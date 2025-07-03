package metrics

type Metrics struct {
	NodeMetrics *NodeMetrics  `json:"node_metrics,omitempty"`
	PodMetrics  []*PodMetrics `json:"pod_metrics,omitempty"`
}
