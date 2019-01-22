// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type networkCryptoServiceServer struct {
	Network cryptoservices.Network
}

func (s networkCryptoServiceServer) JoinRequestMIC(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s networkCryptoServiceServer) JoinAcceptMIC(ctx context.Context, req *ttnpb.JoinAcceptMICRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s networkCryptoServiceServer) EncryptJoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s networkCryptoServiceServer) EncryptRejoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s networkCryptoServiceServer) DeriveNwkSKeys(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.NwkSKeysResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s networkCryptoServiceServer) NwkKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}
