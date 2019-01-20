// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type applicationCryptoServiceServer struct {
	Application cryptoservices.Application
}

func (s applicationCryptoServiceServer) DeriveAppSKey(context.Context, *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	return nil, nil
}

func (s applicationCryptoServiceServer) AppKey(context.Context, *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	return nil, nil
}
