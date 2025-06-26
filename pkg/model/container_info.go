package model

import (
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type ContainerInfo struct {
	Container      *runtimeapi.Container
	ContainerStats *runtimeapi.ContainerStats
}
