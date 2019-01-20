// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type networkCryptoServiceServer struct {
	Network cryptoservices.Network
}

func (s networkCryptoServiceServer) JoinRequestMIC(context.Context, *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	return nil, nil
}

func (s networkCryptoServiceServer) JoinAcceptMIC(context.Context, *ttnpb.JoinAcceptMICRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	return nil, nil
}

func (s networkCryptoServiceServer) EncryptJoinAccept(context.Context, *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	return nil, nil
}

func (s networkCryptoServiceServer) EncryptRejoinAccept(context.Context, *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	return nil, nil
}

func (s networkCryptoServiceServer) DeriveNwkSKeys(context.Context, *ttnpb.DeriveSessionKeysRequest) (*ttnpb.NwkSKeysResponse, error) {
	return nil, nil
}

func (s networkCryptoServiceServer) NwkKey(context.Context, *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	return nil, nil
}
