package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func FindProcessPID(binary string) (int, error) {
	entries, _ := os.ReadDir("/proc")

	for _, e := range entries {
		if !e.IsDir() || !isNumeric(e.Name()) {
			continue
		}

		content, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil {
			continue
		}

		if strings.Contains(string(content), binary) {
			pid, _ := strconv.Atoi(e.Name())
			return pid, nil
		}
	}

	return 0, fmt.Errorf("process %s not found", binary)
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
