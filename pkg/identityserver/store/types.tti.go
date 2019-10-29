// Copyright © 2019 The Things Industries B.V.

package store

import (
	"github.com/gogo/protobuf/types"
)

// WrappedUint64 adds methods to uint64 to allow conversion to types.UInt64value.
type WrappedUint64 uint64

func wrappedUint64(v *types.UInt64Value) *WrappedUint64 {
	if v == nil {
		return nil
	}
	converted := WrappedUint64(v.Value)
	return &converted
}

func (v *WrappedUint64) toPB() *types.UInt64Value {
	if v == nil {
		return nil
	}
	converted := types.UInt64Value{Value: uint64(*v)}
	return &converted
}
