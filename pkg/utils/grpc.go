package utils

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AcquireGrpcConnection establishes a gRPC connection to the given UNIX socket.
// It returns an open *grpc.ClientConn. If the connection fails, the program terminates with a fatal log.
// The caller is responsible for closing the connection.
func AcquireGrpcConnection(socket string, logger *zap.Logger) *grpc.ClientConn {
	connection, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {

		logger.Fatal("failed to connect to gRPC socket", zap.Error(err))
	}
	return connection
}
