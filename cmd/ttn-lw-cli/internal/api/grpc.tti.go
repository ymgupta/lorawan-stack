// Copyright © 2019 The Things Industries B.V.

package api

import (
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"google.golang.org/grpc"
)

// GetCredentials gets per-RPC credentials, optionally applying the given overrides.
// The overrides occur only if both are present.
func GetCredentials(overrideAuthType, overrideAuthValue string) grpc.CallOption {
	if auth == nil {
		return grpc.PerRPCCredentials(rpcmetadata.MD{
			AuthType:  overrideAuthType,
			AuthValue: overrideAuthValue,
		})
	}
	md := *auth
	if overrideAuthType != "" && overrideAuthValue != "" {
		md.AuthType = overrideAuthType
		md.AuthValue = overrideAuthValue
	}
	return grpc.PerRPCCredentials(md)
}
