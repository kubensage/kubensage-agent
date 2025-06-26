package discovery

import (
	"github.com/kubensage/kubensage-agent/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log"
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
		err := conn.Close()
		if err != nil {
			log.Printf("Failed to close client connection: %v", err)
		}
	}(conn)

	runtimeClient := runtimeapi.NewRuntimeServiceClient(conn)

	podSandboxes, err := PodDiscovery(runtimeClient)

	if err != nil {
		return err
	}

	for _, sandbox := range podSandboxes {
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

		log.Printf("PodInfo: %s", jsonStr)
	}

	return nil
}
