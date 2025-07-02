package discovery

import (
	"fmt"
	"log"
	"os"
)

// criSocketCandidates is a map that associates CRI runtime names to a list of possible socket file paths.
// This map is used to check which runtime is in use by verifying the existence of the respective socket files.
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

// CriSocketDiscovery attempts to detect the CRI (Container Runtime Interface) socket file
// by checking known socket file paths for containerd, crio, and dockershim runtimes.
// It returns the URI of the detected socket or an error if no known CRI socket is found.
func CriSocketDiscovery() (string, error) {
	// Iterate over each runtime and its associated socket paths
	for runtime, paths := range criSocketCandidates {
		// Check each socket path for existence and validity
		for _, p := range paths {
			// Check if the path exists and is not a directory
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				// Log the detection of the runtime and the socket path
				log.Printf("Detected CRI runtime: %s (socket: %s)", runtime, p)
				// Return the URI of the socket in the form of "unix://<path>"
				return "unix://" + p, nil
			}
		}
	}

	// Return an error if no known CRI socket files were found
	return "", fmt.Errorf("no known CRI sockets found")
}
