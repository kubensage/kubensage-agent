package utils

import (
	kgrpc "github.com/kubensage/go-common/grpc"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func SetupCRIConnection(
	socket string,
	logger *zap.Logger,
) (client cri.RuntimeServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to CRI socket", zap.String("socket", socket))
	conn := kgrpc.InsecureGrpcConnection(socket, logger)
	logger.Info("Connected to CRI socket")
	return cri.NewRuntimeServiceClient(conn), conn
}

func SetupRelayConnection(
	addr string,
	logger *zap.Logger,
) (client gen.MetricsServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := kgrpc.InsecureGrpcConnection(addr, logger)
	logger.Info("Connected to relay GRPC server")
	return gen.NewMetricsServiceClient(conn), conn
}
