// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type applicationCryptoServiceServer struct {
	Application cryptoservices.Application
}

func (s applicationCryptoServiceServer) DeriveAppSKey(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s applicationCryptoServiceServer) AppKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}
