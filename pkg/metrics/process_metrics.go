package metrics

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProcessMetrics contains various metrics for a process
type ProcessMetrics struct {
	PID           int       // Process ID
	Name          string    // Process name
	Cmdline       string    // Full command line
	Uptime        float64   // Process uptime
	CPUUserTime   float64   // CPU time in user mode (seconds)
	CPUSystemTime float64   // CPU time in kernel mode (seconds)
	CPUPercent    float64   // Estimated CPU usage percentage (optional)
	MemoryRSS     uint64    // Resident Set Size (bytes)
	MemoryVirtual uint64    // Virtual Memory Size (bytes)
	NumThreads    int       // Number of active threads
	OpenFileCount int       // Number of open file descriptors
	ReadBytes     uint64    // Bytes read from disk
	WriteBytes    uint64    // Bytes written to disk
	StartTime     time.Time // Process start time (calculated)
	IsRunning     bool      // Flag indicating if the process is running
}

// ReadMetrics reads the metrics of a given process by its PID
func ReadMetrics(pid int) (ProcessMetrics, error) {
	var processMetrics ProcessMetrics
	processMetrics.PID = pid

	// name
	name, err := GetProcessName(pid)
	if err != nil {
		return processMetrics, err
	}
	processMetrics.Name = name

	// cmdline
	cmdline, err := getCmdline(pid)
	if err != nil {
		return processMetrics, err
	}
	processMetrics.Cmdline = cmdline

	// uptime
	uptime, err := getProcessUptime(pid)
	if err != nil {
		return processMetrics, err
	}
	processMetrics.Uptime = uptime

	// isRunning
	processMetrics.IsRunning = checkProcessRunning(pid)

	return processMetrics, nil
}

// ToJsonString serializes the ProcessMetrics struct to a JSON string
func ToJsonString(pm ProcessMetrics) (string, error) {
	jsonData, err := json.MarshalIndent(pm, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal process metrics to JSON: %v", err)
	}
	return string(jsonData), nil
}
