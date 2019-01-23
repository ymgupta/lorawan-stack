// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type applicationCryptoServiceServer struct {
	Application    cryptoservices.Application
	KeyVault       crypto.KeyVault
	ExposeRootKeys bool
}

func (s applicationCryptoServiceServer) DeriveAppSKey(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	appSKey, err := s.Application.DeriveAppSKey(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	res := &ttnpb.AppSKeyResponse{}
	res.AppSKey, err = cryptoutil.WrapAES128Key(appSKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s applicationCryptoServiceServer) AppKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	if s.Application == nil {
		return nil, errServiceNotSupported
	}
	if !s.ExposeRootKeys {
		return nil, errRootKeysNotExposed
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	appKey, err := s.Application.AppKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	env, err := cryptoutil.WrapAES128Key(appKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
