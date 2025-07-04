package utils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func AcquireGrpcConnection(socket string) *grpc.ClientConn {
	connection, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}(connection)

	return connection
}
