package utils

import (
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// AcquireGrpcConnection establishes a gRPC connection to the given UNIX socket.
// It returns an open *grpc.ClientConn. If the connection fails, the program terminates with a fatal log.
// The caller is responsible for closing the connection.
func acquireGrpcConnection(socket string, logger *zap.Logger) *grpc.ClientConn {
	connection, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {

		logger.Fatal("failed to connect to gRPC socket", zap.Error(err))
	}
	return connection
}

func SetupCRIConnection(
	socket string,
	logger *zap.Logger,
) (client cri.RuntimeServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to CRI socket", zap.String("socket", socket))
	conn := acquireGrpcConnection(socket, logger)
	logger.Info("Connected to CRI socket")
	return cri.NewRuntimeServiceClient(conn), conn
}

func SetupRelayConnection(
	addr string,
	logger *zap.Logger,
) (client proto.MetricsServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := acquireGrpcConnection(addr, logger)
	logger.Info("Connected to relay GRPC server")
	return proto.NewMetricsServiceClient(conn), conn
}
