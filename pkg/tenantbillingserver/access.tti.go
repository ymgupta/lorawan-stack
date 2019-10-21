// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"google.golang.org/grpc"
)

const (
	// BillingReporterAuthType is the AuthType used for tenant billing reporting.
	BillingReporterAuthType = "BillingReporterKey"
)

var (
	errUnauthenticated           = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errNoBillingRights           = errors.DefinePermissionDenied("no_billing_rights", "no billing rights")
	errInvalidBillingReporterKey = errors.DefinePermissionDenied("billing_reporter_key", "invalid billing reporter key")
	errUnsupportedAuthorization  = errors.DefineUnauthenticated("unsupported_authorization", "unsupported authorization method")
)

type billingRightsKeyType struct{}

var billingRightsKey billingRightsKeyType

type billingRights struct {
	report bool
}

func billingRightsFromContext(ctx context.Context) billingRights {
	if rights, ok := ctx.Value(billingRightsKey).(billingRights); ok {
		return rights
	}
	return billingRights{}
}

func newContextWithBillingRights(parent context.Context, rights billingRights) context.Context {
	return context.WithValue(parent, billingRightsKey, rights)
}

func (tbs *TenantBillingServer) billingRightsHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		md := rpcmetadata.FromIncomingContext(ctx)
		if md.AuthType == "" {
			return nil, errUnauthenticated
		}
		rights := billingRights{}
		switch strings.ToLower(md.AuthType) {
		case strings.ToLower(BillingReporterAuthType):
			key, err := hex.DecodeString(md.AuthValue)
			if err != nil {
				return nil, errInvalidBillingReporterKey.WithCause(err)
			}
			var isValidKey bool
			for _, acceptedKey := range tbs.config.decodedReporterKeys {
				if subtle.ConstantTimeCompare(acceptedKey, key) == 1 {
					isValidKey = true
				}
			}
			if !isValidKey {
				return nil, errInvalidBillingReporterKey
			}
			rights.report = true
		default:
			return nil, errUnsupportedAuthorization
		}
		return h(newContextWithBillingRights(ctx, rights), req)
	}
}
