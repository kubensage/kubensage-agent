package metrics

type PodMetrics struct {
	Id        string `json:"id,omitempty"`
	Uid       string `json:"uid,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	State     string `json:"state,omitempty"`
	Attempt   uint32 `json:"attempt,omitempty"`

	ContainerMetrics []*ContainerMetrics `json:"container_metrics,omitempty"`
}

type ContainerMetrics struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Image     string `json:"image,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	State     string `json:"state,omitempty"`
	Attempt   uint32 `json:"attempt,omitempty"`

	CpuMetrics        CpuMetrics        `json:"cpu_metrics,omitempty"`
	MemoryMetrics     MemoryMetrics     `json:"memory_metrics,omitempty"`
	FileSystemMetrics FileSystemMetrics `json:"file_system_metrics,omitempty"`
	SwapMetrics       SwapMetrics       `json:"swap_metrics,omitempty"`
}
