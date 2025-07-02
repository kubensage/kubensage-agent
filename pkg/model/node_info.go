package model

import "github.com/shirou/gopsutil/v3/net"

type NodeInfo struct {
	Hostname        string
	Uptime          uint64
	BootTime        uint64
	Procs           uint64
	OS              string
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	KernelVersion   string
	KernelArch      string
	HostID          string

	CPUModel        string
	CPUCores        int32
	CPUUsagePercent float64

	TotalMemory    uint64
	FreeMemory     uint64
	UsedMemory     uint64
	MemoryUsedPerc float64

	NetworkInterfaces []net.InterfaceStat
}
