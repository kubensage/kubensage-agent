package utils

import (
	proto "github.com/kubensage/kubensage-agent/proto/gen"
	grpc2 "gitlab.com/kubensage/go-common/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func SetupCRIConnection(
	socket string,
	logger *zap.Logger,
) (client cri.RuntimeServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to CRI socket", zap.String("socket", socket))
	conn := grpc2.InsecureGrpcConnection(socket, logger)
	logger.Info("Connected to CRI socket")
	return cri.NewRuntimeServiceClient(conn), conn
}

func SetupRelayConnection(
	addr string,
	logger *zap.Logger,
) (client proto.MetricsServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := grpc2.InsecureGrpcConnection(addr, logger)
	logger.Info("Connected to relay GRPC server")
	return proto.NewMetricsServiceClient(conn), conn
}
