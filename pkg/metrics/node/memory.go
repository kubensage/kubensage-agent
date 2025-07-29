package node

import (
	"context"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
	"sort"
)

// listTopMem collects memory usage information for all running processes and returns
// the top N processes ranked by their Resident Set Size (RSS) in descending order.
//
// It uses gopsutil to access process information, skipping any processes for which
// memory or name data is unavailable. The result includes PID, process name, and memory
// usage in bytes.
//
// Parameters:
//
//   - ctx context.Context:
//     The context used to control cancellation and deadlines for process data retrieval.
//
//   - topN int:
//     The number of top memory-consuming processes to return. If fewer than topN processes
//     are available or accessible, the result may contain fewer entries.
//
//   - logger *zap.Logger:
//     Logger used for debug tracing throughout the execution.
//
// Returns:
//
//   - []*gen.ProcessMemInfo:
//     A slice of ProcessMemInfo entries sorted by memory usage (RSS) in descending order.
//     Each entry includes the process PID, name, and memory usage in bytes.
//
//   - error:
//     Returns an error if the initial retrieval of processes fails. Errors during
//     per-process inspection (e.g., memory or name access) are silently skipped.
func listTopMem(
	ctx context.Context,
	topN int,
	logger *zap.Logger,
) ([]*gen.ProcessMemInfo, error) {
	logger.Debug("Start listTopMem")

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

	logger.Debug("Finish listTopMem")

	return topProcesses, nil
}
