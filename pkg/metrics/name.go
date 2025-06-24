package metrics

import (
	"fmt"
	"os"
	"strings"
)

// GetProcessName retrieves the name of the process from /proc/[pid]/status
func GetProcessName(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/%d/status: %v", pid, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Name:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				return fields[1], nil
			}
		}
	}
	return "", fmt.Errorf("process name not found for pid %d", pid)
}
