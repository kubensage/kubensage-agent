package utils

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func ConvertCRIUInt64(v *runtimeapi.UInt64Value) *wrapperspb.UInt64Value {
	if v == nil {
		return nil
	}
	return &wrapperspb.UInt64Value{Value: v.Value}
}
