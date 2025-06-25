package main

import "github.com/kubensage/kubensage-agent/pkg/discovery"

func main() {
	discovery.A()
	/*targets := []string{"kube-apiserver", "kubelet", "etcd", "kube-controller-manager", "kube-scheduler"}

	for _, name := range targets {
		// Trova il PID del processo
		pid, err := discovery.FindProcessPID(name)
		if err != nil {
			log.Printf("Skipping %s: %v", name, err)
			continue
		}

		// Legge le metriche del processo
		processMetrics, err := metrics.ReadMetrics(pid)
		if err != nil {
			log.Printf("Failed to read metrics for %s (pid %d): %v", name, pid, err)
			continue
		}

		// Converte le metriche in JSON
		jsonStr, err := metrics.ToJsonString(processMetrics)
		if err != nil {
			log.Printf("Failed to serialize metrics for %s (pid %d) to JSON: %v", name, pid, err)
			continue
		}

		// Log delle metriche del processo
		log.Printf("Metrics for %s (pid %d): %s", name, pid, jsonStr)
	}*/
}
