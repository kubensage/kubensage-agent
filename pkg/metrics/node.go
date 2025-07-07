package metrics

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"time"
)

// NodeMetrics represents a full set of system-level metrics for a Kubernetes node.
// It includes metadata, CPU/memory usage, PSI (Pressure Stall Information), and network interfaces.
type NodeMetrics struct {
	// Basic host metadata
	Hostname        string // Hostname of the node
	Uptime          uint64 // Uptime in seconds
	BootTime        uint64 // Boot time (Unix timestamp)
	Procs           uint64 // Number of running processes
	OS              string // OS name (e.g., "linux")
	Platform        string // Distribution name (e.g., "ubuntu")
	PlatformFamily  string // OS family (e.g., "debian")
	PlatformVersion string // OS version (e.g., "25.04")
	KernelVersion   string // Kernel version (e.g., "6.14.0-22-generic")
	KernelArch      string // CPU architecture (e.g., "x86_64")
	HostID          string // Machine ID (UUID or host ID)

	// CPU metrics
	CPUModel        string  // CPU model name
	CPUCores        int32   // Number of logical CPU cores
	CPUUsagePercent float64 // Total CPU usage over a 1-second window

	// Memory metrics
	TotalMemory    uint64  // Total physical memory in bytes
	FreeMemory     uint64  // Free memory in bytes
	UsedMemory     uint64  // Used memory in bytes
	MemoryUsedPerc float64 // Used memory as a percentage of total

	// PSI metrics (Pressure Stall Information from /proc/pressure)
	PsiCpuMetrics    PsiMetrics // CPU pressure stall data
	PsiMemoryMetrics PsiMetrics // Memory pressure stall data
	PsiIoMetrics     PsiMetrics // I/O pressure stall data

	// Network interfaces (excluding stats like RX/TX bytes)
	NetworkInterfaces []net.InterfaceStat // List of detected network interfaces
}

// SafeNodeMetrics collects node-level system metrics using gopsutil and /proc/pressure.
// It gathers host metadata, CPU/memory usage, PSI metrics, and network interface details.
// Returns a NodeMetrics struct on success or an error if any critical system call fails.
func SafeNodeMetrics(ctx context.Context, interval time.Duration, logger zap.Logger) (*NodeMetrics, error) {
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

		PsiCpuMetrics:    SafePsiMetrics("/proc/pressure/cpu", logger),
		PsiMemoryMetrics: SafePsiMetrics("/proc/pressure/memory", logger),
		PsiIoMetrics:     SafePsiMetrics("/proc/pressure/io", logger),

		NetworkInterfaces: netInfo,
	}

	return nodeInfo, nil
}
