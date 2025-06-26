package model

import (
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type PodInfo struct {
	Timestamp  int64
	Pod        *runtimeapi.PodSandbox
	Containers []*ContainerInfo
}
