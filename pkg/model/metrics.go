package model

import "github.com/shirou/gopsutil/v3/net"

type Metrics struct {
	NodeMetrics NodeMetrics  `json:"node_metrics,omitempty"`
	PodMetrics  []PodMetrics `json:"pod_metrics,omitempty"`
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

	ContainerMetrics []ContainerMetrics `json:"container_metrics,omitempty"`
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
	IOMetrics         IOMetrics         `json:"io_metrics,omitempty"`
}

type CpuMetrics struct {
	Timestamp            int64      `json:"timestamp,omitempty"`
	UsageCoreNanoSeconds uint64     `json:"usage_core_nano_seconds,omitempty"`
	UsageNanoCores       uint64     `json:"usage_nano_cores,omitempty"`
	PsiSome              PsiMetrics `json:"psi_some_metrics,omitempty"`
	PsiFull              PsiMetrics `json:"psi_full_metrics,omitempty"`
}

type MemoryMetrics struct {
	Timestamp       int64      `json:"timestamp,omitempty"`
	WorkingSetBytes uint64     `json:"working_set_bytes,omitempty"`
	AvailableBytes  uint64     `json:"available_bytes,omitempty"`
	UsageBytes      uint64     `json:"usage_bytes,omitempty"`
	RssBytes        uint64     `json:"rss_bytes,omitempty"`
	PageFaults      uint64     `json:"page_faults,omitempty"`
	MajorPageFaults uint64     `json:"major_page_faults,omitempty"`
	PsiSome         PsiMetrics `json:"psi_some_metrics,omitempty"`
	PsiFull         PsiMetrics `json:"psi_full_metrics,omitempty"`
}

type PsiMetrics struct {
	Total  uint64  `json:"Total,omitempty"`
	Avg10  float64 `json:"Avg10,omitempty"`
	Avg60  float64 `json:"Avg60,omitempty"`
	Avg300 float64 `json:"Avg300,omitempty"`
}

type FileSystemMetrics struct {
	Timestamp  int64  `json:"timestamp,omitempty"`
	Mountpoint string `json:"mountpoint,omitempty"`
	UsedBytes  uint64 `json:"used_bytes,omitempty"`
	InodesUsed uint64 `json:"inodes_used,omitempty"`
}

type SwapMetrics struct {
	Timestamp          int64  `json:"timestamp,omitempty"`
	SwapAvailableBytes uint64 `json:"swap_available_bytes,omitempty"`
	SwapUsageBytes     uint64 `json:"swap_usage_bytes,omitempty"`
}

type IOMetrics struct {
	PsiSome PsiMetrics `json:"psi_some_metrics,omitempty"`
	PsiFull PsiMetrics `json:"psi_full_metrics,omitempty"`
}
