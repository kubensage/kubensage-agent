package utils

import (
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// SafeUint64ValueToInt64OrDefault safely extracts the int64 value from a *runtimeapi.UInt64Value.
// If the input is nil, it returns the provided default value instead.
// This is useful for handling optional protobuf fields.
func SafeUint64ValueToInt64OrDefault(
	field *runtimeapi.UInt64Value,
	defaultVal int64,
) int64 {
	if field != nil {
		return int64(field.Value)
	}
	return defaultVal
}
