package node

import (
	"context"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/process"
	"sort"
)

// topMem returns the top N memory-consuming processes currently running on the system.
// It uses the gopsutil library to retrieve memory usage (RSS) and process names.
//
// Parameters:
//   - ctx:   context for handling timeouts or cancellations during system calls
//   - topN:  the number of top memory-using processes to return
//
// Returns:
//   - A slice of *gen.ProcessMemInfo containing up to topN processes sorted by RSS in descending order
//   - An error if the initial process list retrieval fails (individual process failures are skipped)
//
// Notes:
//   - Processes that fail to report memory or name are skipped silently.
//   - If fewer than topN valid processes are available, the result will contain less than topN items.
func topMem(ctx context.Context, topN int) ([]*gen.ProcessMemInfo, error) {
	// Retrieve the list of all running processes
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Preallocate memory for efficiency based on the total number of processes
	processesMemInfo := make([]*gen.ProcessMemInfo, 0, len(processes))

	// Collect memory usage and name for each valid process
	for _, p := range processes {
		memInfo, err := p.MemoryInfoWithContext(ctx)
		if err != nil {
			continue // Skip processes with unreadable memory info
		}
		name, err := p.NameWithContext(ctx)
		if err != nil {
			continue // Skip processes with unreadable names
		}
		processesMemInfo = append(processesMemInfo, &gen.ProcessMemInfo{
			Pid:    p.Pid,
			Name:   name,
			Memory: memInfo.RSS, // Resident Set Size (actual memory in use)
		})
	}

	// Sort all valid processes by memory usage in descending order
	sort.Slice(processesMemInfo, func(i, j int) bool {
		return processesMemInfo[i].Memory > processesMemInfo[j].Memory
	})

	// Limit the result to topN items or less if fewer are available
	top := topN
	if len(processesMemInfo) < topN {
		top = len(processesMemInfo)
	}
	topProcesses := make([]*gen.ProcessMemInfo, top)
	copy(topProcesses, processesMemInfo[:top])

	return topProcesses, nil
}
