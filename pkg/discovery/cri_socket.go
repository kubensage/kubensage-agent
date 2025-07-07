package discovery

import (
	"fmt"
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
func CriSocketDiscovery() (string, error) {
	for _, paths := range criSocketCandidates {
		for _, p := range paths {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return "unix://" + p, nil
			}
		}
	}
	return "", fmt.Errorf("no known CRI sockets found")
}
