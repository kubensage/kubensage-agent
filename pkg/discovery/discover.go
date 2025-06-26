package discovery

import (
	"fmt"
	"github.com/kubensage/kubensage-agent/pkg/model"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"google.golang.org/grpc"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"time"
)

func Discover() ([]model.PodInfo, error) {
	// Discover the socket
	socket, err := criSocketDiscovery()
	if err != nil {
		return nil, fmt.Errorf("failed to discover CRI socket: %v", err)
	}

	// Create the gRPC connection
	conn, err := utils.GrpcClientConnection(socket)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close client connection: %v", err)
		}
	}(conn)

	// Create the runtime client
	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)

	// Discover the pods
	podSandboxes, err := listPods(runtimeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %v", err)
	}

	var allPodInfo []model.PodInfo

	// Iterate through the pods
	for _, sandbox := range podSandboxes {
		podInfo := model.PodInfo{Timestamp: time.Now().UnixNano(), Pod: sandbox}

		// Retrieve the pod stats
		podStats, err := listPodStats(runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to list pod stats for sandbox %s: %v", sandbox.Id, err)
		} else {
			podInfo.PodStats = podStats
		}

		// Retrieve the containers of the pod
		containers, err := listContainers(runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to discover containers for sandbox %s: %v", sandbox.Id, err)
			continue
		}

		var containerInfos []*model.ContainerInfo

		// Retrieve the stats for each container
		for _, container := range containers {
			stats, err := listContainerStats(runtimeClient, sandbox.Id, container.Id)
			if err != nil {
				log.Printf("Failed to discover stats for container %s: %v", container.Id, err)
				continue
			}

			containerInfo := &model.ContainerInfo{
				Container:      container,
				ContainerStats: stats,
			}

			containerInfos = append(containerInfos, containerInfo)
		}

		podInfo.Containers = containerInfos

		allPodInfo = append(allPodInfo, podInfo)
	}

	return allPodInfo, nil
}
