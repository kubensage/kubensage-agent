package discovery

import (
	"context"
	"fmt"
	"github.com/kubensage/kubensage-agent/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func A() {
	conn, err := grpc.NewClient("unix:///var/run/containerd/containerd.sock",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Printf("Failed to connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Failed to close client connection: %v", err)
		}
	}(conn)

	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Lista container
	resp, err := runtimeClient.ListPodSandbox(ctx, &runtimeapi.ListPodSandboxRequest{})

	if err != nil {
		log.Printf("Failed to list pod sandboxes: %v", err)
	}

	fmt.Printf("Found %d pod sandboxes:\n", len(resp.Items))

	for _, sandbox := range resp.Items {
		podInfo := model.PodInfo{
			Id:          sandbox.Id,
			Name:        sandbox.Metadata.Name,
			Namespace:   sandbox.Metadata.Namespace,
			Uid:         sandbox.Metadata.Uid,
			State:       sandbox.State.String(),
			CreatedAt:   sandbox.CreatedAt,
			Annotations: sandbox.Annotations,
			Labels:      sandbox.Labels,
		}

		jsonStr, err := model.ToJsonString(podInfo)
		if err != nil {
			log.Printf("Failed to serialize PodInfo for sandbox %s: %v", sandbox.Id, err)
			continue
		}

		stats, err := runtimeClient.PodSandboxStats(ctx, &runtimeapi.PodSandboxStatsRequest{PodSandboxId: sandbox.Id})
		log.Printf("Stats: %s", stats)

		log.Printf("PodInfo: %s", jsonStr)
	}

	/*resp, err := runtimeClient.ListContainers(ctx, &runtimeapi.ListContainersRequest{})
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
	}*/
}

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
