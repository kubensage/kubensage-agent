package metrics

import (
	"fmt"
	"os"
	"strings"
)

// getCmdline legge la command line completa di un processo dal file /proc/[pid]/cmdline
func getCmdline(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/%d/cmdline: %v", pid, err)
	}

	// Gli argomenti sono separati da \0, che non sono visibili in una stringa normale
	args := strings.Split(string(data), "\x00")

	// Rimuove eventuali stringhe vuote (specialmente l'ultima)
	var cleaned []string
	for _, arg := range args {
		if arg != "" {
			cleaned = append(cleaned, arg)
		}
	}

	return strings.Join(cleaned, " "), nil
}
