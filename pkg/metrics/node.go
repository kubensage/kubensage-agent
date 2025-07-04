package metrics

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

// NodeMetrics represents a full set of system-level metrics for a Kubernetes node.
// It includes metadata, CPU/memory usage, PSI (Pressure Stall Information), and network interfaces.
type NodeMetrics struct {
	// Basic host metadata
	Hostname        string `json:"hostname,omitempty"`         // Hostname of the node
	Uptime          uint64 `json:"uptime,omitempty"`           // Uptime in seconds
	BootTime        uint64 `json:"boot_time,omitempty"`        // Boot time (Unix timestamp)
	Procs           uint64 `json:"procs,omitempty"`            // Number of running processes
	OS              string `json:"os,omitempty"`               // OS name (e.g., "linux")
	Platform        string `json:"platform,omitempty"`         // Distribution name (e.g., "ubuntu")
	PlatformFamily  string `json:"platform_family,omitempty"`  // OS family (e.g., "debian")
	PlatformVersion string `json:"platform_version,omitempty"` // OS version (e.g., "25.04")
	KernelVersion   string `json:"kernel_version,omitempty"`   // Kernel version (e.g., "6.14.0-22-generic")
	KernelArch      string `json:"kernel_arch,omitempty"`      // CPU architecture (e.g., "x86_64")
	HostID          string `json:"host_id,omitempty"`          // Machine ID (UUID or host ID)

	// CPU metrics
	CPUModel        string  `json:"cpu_model,omitempty"`         // CPU model name
	CPUCores        int32   `json:"cpu_cores,omitempty"`         // Number of logical CPU cores
	CPUUsagePercent float64 `json:"cpu_usage_percent,omitempty"` // Total CPU usage over a 1-second window

	// Memory metrics
	TotalMemory    uint64  `json:"total_memory,omitempty"`     // Total physical memory in bytes
	FreeMemory     uint64  `json:"free_memory,omitempty"`      // Free memory in bytes
	UsedMemory     uint64  `json:"used_memory,omitempty"`      // Used memory in bytes
	MemoryUsedPerc float64 `json:"memory_used_perc,omitempty"` // Used memory as a percentage of total

	// PSI metrics (Pressure Stall Information from /proc/pressure)
	PsiCpuMetrics    PsiMetrics `json:"psi_cpu_metrics,omitempty"`    // CPU pressure stall data
	PsiMemoryMetrics PsiMetrics `json:"psi_memory_metrics,omitempty"` // Memory pressure stall data
	PsiIoMetrics     PsiMetrics `json:"psi_io_metrics,omitempty"`     // I/O pressure stall data

	// Network interfaces (excluding stats like RX/TX bytes)
	NetworkInterfaces []net.InterfaceStat `json:"network_interfaces,omitempty"` // List of detected network interfaces
}

// SafeNodeMetrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func SafeNodeMetrics(ctx context.Context, interval time.Duration) (*NodeMetrics, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get CPU usage over 1 second interval (non-per-core)
	cpuPercent, err := cpu.PercentWithContext(ctx, interval, false)

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	netInfo, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	nodeInfo := &NodeMetrics{
		Hostname:        info.Hostname,
		Uptime:          info.Uptime,
		BootTime:        info.BootTime,
		Procs:           info.Procs,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		HostID:          info.HostID,

		CPUModel:        cpuInfo[0].ModelName,
		CPUCores:        cpuInfo[0].Cores,
		CPUUsagePercent: cpuPercent[0],

		TotalMemory:    memInfo.Total,
		FreeMemory:     memInfo.Free,
		UsedMemory:     memInfo.Used,
		MemoryUsedPerc: memInfo.UsedPercent,

		PsiCpuMetrics:    SafePsiMetrics("/proc/pressure/cpu"),
		PsiMemoryMetrics: SafePsiMetrics("/proc/pressure/memory"),
		PsiIoMetrics:     SafePsiMetrics("/proc/pressure/io"),

		NetworkInterfaces: netInfo,
	}

	return nodeInfo, nil
}
