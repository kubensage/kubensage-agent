package metrics

// Metrics is the top-level structure that holds all collected telemetry data for a node.
// It includes both node-level metrics (e.g., CPU, memory, network, PSI) and a list of pod-level metrics.
// This structure is typically serialized and sent to a relay/exporter.
type Metrics struct {
	NodeMetrics *NodeMetrics  // Host-level metrics
	PodMetrics  []*PodMetrics // Metrics grouped per pod and container
}
