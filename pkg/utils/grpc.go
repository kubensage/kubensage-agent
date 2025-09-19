package utils

import (
	"github.com/kubensage/go-common/grpc"
	"github.com/kubensage/kubensage-agent/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SetupCRIConnection establishes a gRPC connection to the Container Runtime Interface (CRI)
// runtime service via a Unix domain socket.
//
// This function uses an insecure gRPC dial and logs connection status.
// It returns both the typed client and the raw gRPC connection so the caller
// can later close the connection properly.
//
// Parameters:
//   - socket string:
//     The Unix socket path to the CRI runtime (e.g., "unix:///var/run/containerd/containerd.sock").
//   - logger *zap.Logger:
//     Logger used to log connection attempts and results.
//
// Returns:
//   - cri.RuntimeServiceClient:
//     A gRPC client to interact with the CRI runtime.
//   - *grpc.ClientConn:
//     The underlying gRPC connection that must be closed by the caller when no longer needed.
func SetupCRIConnection(
	socket string,
	logger *zap.Logger,
) (client cri.RuntimeServiceClient, connection *grpc.ClientConn) {
	logger.Info("connecting to CRI socket", zap.String("socket", socket))
	conn := gogrpc.InsecureGrpcConnection(socket, logger)
	logger.Info("connected to CRI socket")
	return cri.NewRuntimeServiceClient(conn), conn
}

// SetupRelayConnection establishes a gRPC connection to the relay metrics service,
// which is responsible for receiving node and container metrics.
//
// This function uses an insecure gRPC dial and logs connection status.
// It returns both the typed client and the raw gRPC connection so the caller
// can later close the connection properly.
//
// Parameters:
//   - addr string:
//     The network address (host:port) of the relay service.
//   - logger *zap.Logger:
//     Logger used to log connection attempts and results.
//
// Returns:
//   - gen.MetricsServiceClient:
//     A gRPC client to send metrics to the relay service.
//   - *grpc.ClientConn:
//     The underlying gRPC connection that must be closed by the caller when no longer needed.
func SetupRelayConnection(
	addr string,
	logger *zap.Logger,
) (client gen.MetricsServiceClient, connection *grpc.ClientConn) {
	logger.Info("Connecting to relay GRPC server", zap.String("socket", addr))
	conn := gogrpc.InsecureGrpcConnection(addr, logger)
	logger.Info("Connected to relay GRPC server")
	return gen.NewMetricsServiceClient(conn), conn
}
