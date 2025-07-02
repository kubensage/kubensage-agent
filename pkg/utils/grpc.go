package utils

import (
	"fmt"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GrpcClientConnection(socket string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("cri grpc connection error: %v", types.ErrUnexpectedEndOfGroupEmpty.Error())
	}

	return conn, nil
}
