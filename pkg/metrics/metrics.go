package metrics

import "github.com/shirou/gopsutil/v3/net"

type Metrics struct {
	NodeMetrics *NodeMetrics  `json:"node_metrics,omitempty"`
	PodMetrics  []*PodMetrics `json:"pod_metrics,omitempty"`
}

type NodeMetrics struct {
	Hostname        string `json:"hostname,omitempty"`
	Uptime          uint64 `json:"uptime,omitempty"`
	BootTime        uint64 `json:"boot_time,omitempty"`
	Procs           uint64 `json:"procs,omitempty"`
	OS              string `json:"os,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformFamily  string `json:"platform_family,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	KernelVersion   string `json:"kernel_version,omitempty"`
	KernelArch      string `json:"kernel_arch,omitempty"`
	HostID          string `json:"host_id,omitempty"`

	CPUModel        string  `json:"cpu_model,omitempty"`
	CPUCores        int32   `json:"cpu_cores,omitempty"`
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"`

	TotalMemory    uint64  `json:"total_memory,omitempty"`
	FreeMemory     uint64  `json:"free_memory,omitempty"`
	UsedMemory     uint64  `json:"used_memory,omitempty"`
	MemoryUsedPerc float64 `json:"memory_used_perc,omitempty"`

	NetworkInterfaces []net.InterfaceStat `json:"network_interfaces,omitempty"`
}

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
	IoMetrics         IoMetrics         `json:"io_metrics,omitempty"`
}
