package utils

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// ConvertCRIUInt64 converts a CRI (Container Runtime Interface) UInt64Value to a
// Protobuf wrapperspb.UInt64Value.
//
// This is useful when translating CRI stats (e.g., container memory or CPU usage)
// into a format compatible with Protobuf messages that expect wrapper types
// (which allow distinguishing between zero and nil values).
//
// Parameters:
//   - v *cri.UInt64Value:
//     Pointer to a CRI UInt64Value. May be nil.
//
// Returns:
//   - *wrapperspb.UInt64Value:
//     A new wrapperspb.UInt64Value containing the same numeric value as the input,
//     or nil if the input was nil.
func ConvertCRIUInt64(
	v *cri.UInt64Value,
) *wrapperspb.UInt64Value {
	if v == nil {
		return nil
	}
	return &wrapperspb.UInt64Value{Value: v.Value}
}
