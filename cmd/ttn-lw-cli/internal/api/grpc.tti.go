// Copyright © 2019 The Things Industries B.V.

package api

import "google.golang.org/grpc"

// GetCredentials gets per-RPC credentials, optionally applying the given overrides.
func GetCredentials(overrideAuthType, overrideAuthValue string) grpc.CallOption {
	if auth == nil {
		return nil
	}
	md := *auth
	if overrideAuthType != "" {
		md.AuthType = overrideAuthType
	}
	if overrideAuthValue != "" {
		md.AuthValue = overrideAuthValue
	}
	return grpc.PerRPCCredentials(md)
}
