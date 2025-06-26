package discovery

import (
	"fmt"
	"log"
	"os"
)

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
