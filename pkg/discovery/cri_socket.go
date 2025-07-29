package discovery

import (
	"fmt"
	"os"
)

// criSocketCandidates defines a list of known CRI (Container Runtime Interface) socket paths,
// grouped by runtime type. These are the default installation paths for common runtimes
// like containerd, CRI-O, and dockershim.
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

// CriSocketDiscovery attempts to discover a valid CRI Unix socket on the host.
//
// It iterates over a predefined list of well-known socket paths used by popular
// container runtimes (e.g., containerd, CRI-O, dockershim). The function checks for
// the existence of each socket file and returns the first match found, prefixed with "unix://".
//
// Returns:
//   - string: the full URI to the discovered socket (e.g., "unix:///var/run/containerd/containerd.sock")
//   - error: non-nil if no known CRI sockets were found or accessible
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
