package metrics

import (
	"fmt"
	"os"
	"strings"
)

// getCommandLine reads the full command line of a process from /proc/[pid]/cmdline
func getCmdLine(pid int) (string, error) {
	rawData, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/%d/cmdline: %v", pid, err)
	}

	// Arguments are separated by null bytes (\0), not spaces
	rawArgs := strings.Split(string(rawData), "\x00")

	// Filter out any empty strings (especially at the end)
	var args []string
	for _, arg := range rawArgs {
		if arg != "" {
			args = append(args, arg)
		}
	}

	return strings.Join(args, " "), nil
}
