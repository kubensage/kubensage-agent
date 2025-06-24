package metrics

import (
	"fmt"
	"os"
)

// checkProcessRunning checks if a process is running by checking if its /proc/[pid] directory exists
func checkProcessRunning(pid int) bool {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}
