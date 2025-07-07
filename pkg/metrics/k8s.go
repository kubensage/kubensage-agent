package metrics

// PodMetrics represents a single pod's metadata and its associated container-level metrics.
// It aggregates all container metrics under the pod context.
type PodMetrics struct {
	Id        string // CRI sandbox ID
	Uid       string // Kubernetes UID
	Name      string // Pod name
	Namespace string // Kubernetes namespace
	CreatedAt int64  // Pod creation timestamp (nanoseconds since epoch)
	State     string // Pod state (e.g., SANDBOX_READY, SANDBOX_NOTREADY)
	Attempt   uint32 // Restart attempt count

	ContainerMetrics []*ContainerMetrics `json:"container_metrics,omitempty"` // Metrics for containers in this pod
}

// ContainerMetrics holds detailed resource metrics and metadata for a single container.
// All metrics are safe-to-access and fall back to default values if unavailable.
type ContainerMetrics struct {
	Id        string // CRI container ID
	Name      string // Container name
	Image     string // Container image reference
	CreatedAt int64  // Container creation timestamp (nanoseconds since epoch)
	State     string // Container state (e.g., CONTAINER_RUNNING, CONTAINER_EXITED)
	Attempt   uint32 // Start attempt count

	CpuMetrics        CpuMetrics        // CPU usage statistics
	MemoryMetrics     MemoryMetrics     // Memory usage statistics
	FileSystemMetrics FileSystemMetrics // Filesystem (writable layer) statistics
	SwapMetrics       SwapMetrics       // Swap usage statistics
}
