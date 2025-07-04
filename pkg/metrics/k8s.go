package metrics

// PodMetrics represents a single pod's metadata and its associated container-level metrics.
// It aggregates all container metrics under the pod context.
type PodMetrics struct {
	Id        string `json:"id,omitempty"`         // CRI sandbox ID
	Uid       string `json:"uid,omitempty"`        // Kubernetes UID
	Name      string `json:"name,omitempty"`       // Pod name
	Namespace string `json:"namespace,omitempty"`  // Kubernetes namespace
	CreatedAt int64  `json:"created_at,omitempty"` // Pod creation timestamp (nanoseconds since epoch)
	State     string `json:"state,omitempty"`      // Pod state (e.g., SANDBOX_READY, SANDBOX_NOTREADY)
	Attempt   uint32 `json:"attempt,omitempty"`    // Restart attempt count

	ContainerMetrics []*ContainerMetrics `json:"container_metrics,omitempty"` // Metrics for containers in this pod
}

// ContainerMetrics holds detailed resource metrics and metadata for a single container.
// All metrics are safe-to-access and fall back to default values if unavailable.
type ContainerMetrics struct {
	Id        string `json:"id,omitempty"`         // CRI container ID
	Name      string `json:"name,omitempty"`       // Container name
	Image     string `json:"image,omitempty"`      // Container image reference
	CreatedAt int64  `json:"created_at,omitempty"` // Container creation timestamp (nanoseconds since epoch)
	State     string `json:"state,omitempty"`      // Container state (e.g., CONTAINER_RUNNING, CONTAINER_EXITED)
	Attempt   uint32 `json:"attempt,omitempty"`    // Start attempt count

	CpuMetrics        CpuMetrics        `json:"cpu_metrics,omitempty"`         // CPU usage statistics
	MemoryMetrics     MemoryMetrics     `json:"memory_metrics,omitempty"`      // Memory usage statistics
	FileSystemMetrics FileSystemMetrics `json:"file_system_metrics,omitempty"` // Filesystem (writable layer) statistics
	SwapMetrics       SwapMetrics       `json:"swap_metrics,omitempty"`        // Swap usage statistics
}
