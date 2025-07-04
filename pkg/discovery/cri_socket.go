package discovery

import (
	"fmt"
	"log"
	"os"
)

// criSocketCandidates lists known default CRI socket paths for supported runtimes.
// The keys are runtime names (e.g., containerd, crio) used only for logging.
var criSocketCandidates = map[string][]string{
	"containerd": {
		"/run/containerd/containerd.sock",
		"/var/run/containerd/containerd.sock",
	},
	"crio": {
		"/var/run/crio/crio.sock",
	},
	"dockershim": {
		"/var/run/dockershim.sock",
	},
}

// CriSocketDiscovery attempts to detect the active CRI runtime socket by scanning known paths.
// It returns the full socket URI (e.g., "unix:///run/containerd/containerd.sock") if successful.
// If no known socket is found, it returns an error.
// The function also logs the detected runtime for diagnostic purposes.
func CriSocketDiscovery() (string, error) {
	for runtime, paths := range criSocketCandidates {
		for _, p := range paths {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				log.Printf("Detected CRI runtime: %s (socket: %s)", runtime, p)
				return "unix://" + p, nil
			}
		}
	}
	return "", fmt.Errorf("no known CRI sockets found")
}
