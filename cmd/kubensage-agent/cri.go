package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// This program interacts with any container runtime that implements the Kubernetes Container Runtime Interface (CRI),
// such as containerd, CRI-O, and other CRI-compliant runtimes. It uses gRPC to communicate with the runtime's service
// and retrieve container information, making it compatible with different container runtimes in Kubernetes environments.
func main() {
	// Connessione al socket di containerd
	conn, err := grpc.NewClient("unix:///var/run/containerd/containerd.sock", grpc.WithInsecure())
	if err != nil {
		panic(fmt.Errorf("errore nella connessione al socket: %v", err))
	}
	defer conn.Close()

	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Lista container
	resp, err := runtimeClient.ListContainers(ctx, &runtimeapi.ListContainersRequest{})
	if err != nil {
		panic(fmt.Errorf("errore nella chiamata ListContainers: %v", err))
	}

	fmt.Printf("Trovati %d container:\n\n", len(resp.Containers))

	for _, c := range resp.Containers {
		// ContainerStatus con Verbose=true per ottenere PID
		statusResp, err := runtimeClient.ContainerStatus(ctx, &runtimeapi.ContainerStatusRequest{
			ContainerId: c.Id,
			Verbose:     true,
		})
		if err != nil {
			fmt.Printf("Errore su container %s: %v\n", c.Id, err)
			continue
		}

		labels := statusResp.Status.Labels
		podName := labels["io.kubernetes.pod.name"]
		namespace := labels["io.kubernetes.pod.namespace"]
		containerName := labels["io.kubernetes.container.name"]

		// Ignora container sandbox
		if containerName == "POD" {
			continue
		}

		info := statusResp.Info
		if info == nil {
			fmt.Printf("Nessuna info extra per container %s\n", c.Id)
			continue
		}

		raw, _ := json.Marshal(info)
		var parsed map[string]interface{}
		json.Unmarshal(raw, &parsed)

		innerRaw, ok := parsed["info"].(string)
		if !ok {
			fmt.Printf("Formato info interno non valido per container %s\n", c.Id)
			continue
		}

		var inner map[string]interface{}
		if err := json.Unmarshal([]byte(innerRaw), &inner); err != nil {
			fmt.Printf("Errore decoding JSON interno per container %s: %v\n", c.Id, err)
			continue
		}

		pidFloat, ok := inner["pid"].(float64)
		if !ok || pidFloat == 0 {
			continue // ignora se PID è 0 o non presente
		}
		pid := int(pidFloat)

		// Output identificativo
		fmt.Printf("[%s/%s] Container: %s\n  ID: %s\n  PID: %d\n", namespace, podName, containerName, c.Id[:12], pid)

		// /proc/[pid]/status
		statusPath := fmt.Sprintf("/proc/%d/status", pid)
		if data, err := ioutil.ReadFile(statusPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "VmRSS:") || strings.HasPrefix(line, "Threads:") {
					fmt.Println("   ", line)
				}
			}
		} else {
			if !os.IsNotExist(err) {
				fmt.Println("   Errore lettura /proc:", err)
			}
		}

		fmt.Println()
	}
}
