package utils

import (
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func SafeUint64ValueToInt64OrDefault(field *runtimeapi.UInt64Value, defaultVal int64) int64 {
	if field != nil {
		return int64(field.Value)
	}
	return defaultVal
}
