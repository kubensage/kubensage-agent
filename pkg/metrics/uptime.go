package metrics

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getProcessUptime(pid int) (float64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/%d/stat: %v", pid, err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 22 {
		return 0, errors.New("unexpected format in /proc/[pid]/stat")
	}

	startTimeTicks, err := strconv.ParseUint(fields[21], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid start time: %v", err)
	}

	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/uptime: %v", err)
	}
	uptimeFields := strings.Fields(string(uptimeData))
	systemUptimeSeconds, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid system uptime: %v", err)
	}

	const clockTicks = 100.0

	processUptimeSeconds := systemUptimeSeconds - (float64(startTimeTicks) / clockTicks)
	if processUptimeSeconds < 0 {
		processUptimeSeconds = 0
	}

	return processUptimeSeconds, nil
}
