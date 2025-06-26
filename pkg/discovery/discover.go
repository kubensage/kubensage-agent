package discovery

import (
	"github.com/kubensage/kubensage-agent/pkg/model"
	"github.com/kubensage/kubensage-agent/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
	"time"
)

func Discover() error {
	socket, err := CriSocketDiscovery()

	if err != nil {
		return err
	}

	conn, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close client connection: %v", err)
		}
	}(conn)

	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)

	podSandboxes, err := ListPods(runtimeClient)
	if err != nil {
		return err
	}

	for _, sandbox := range podSandboxes {
		podInfo := model.PodInfo{Timestamp: time.Now().UnixNano(), Pod: sandbox}

		podStats, err := ListPodStats(runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to list pod stats: %v", err)
		} else {
			podInfo.PodStats = podStats
		}

		containers, err := ListContainers(runtimeClient, sandbox.Id)
		if err != nil {
			log.Printf("Failed to discover containers for sandbox %s: %v", sandbox.Id, err)
			continue
		}

		var containerInfos []*model.ContainerInfo

		for _, container := range containers {
			stats, err := ListContainerStats(runtimeClient, sandbox.Id, container.Id)
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

		jsonStr, err := utils.ToJsonString(podInfo)
		if err != nil {
			log.Printf("Failed to serialize PodInfo for sandbox %s: %v", sandbox.Id, err)
			continue
		}

		log.Printf("PodInfo: %s", jsonStr)
	}

	return nil
}
