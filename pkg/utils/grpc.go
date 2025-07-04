package utils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

// AcquireGrpcConnection establishes a gRPC connection to the given UNIX socket.
// It returns an open *grpc.ClientConn. If the connection fails, the program terminates with a fatal log.
// The caller is responsible for closing the connection.
func AcquireGrpcConnection(socket string) *grpc.ClientConn {
	connection, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC socket %q: %v", socket, err)
	}
	return connection
}
