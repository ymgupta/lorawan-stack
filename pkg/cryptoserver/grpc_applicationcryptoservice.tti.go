// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type applicationCryptoServiceServer struct {
	Provisioners Provisioners
	KeyVault     crypto.KeyVault
}

var errNoApplicationService = errors.DefineFailedPrecondition("no_application_service", "no application service provided by provisioner `{id}`")

func (s applicationCryptoServiceServer) DeriveAppSKey(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.AppSKeyResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Application == nil {
		return nil, errNoApplicationService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	appSKey, err := provisioner.Application.DeriveAppSKey(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	res := &ttnpb.AppSKeyResponse{}
	res.AppSKey, err = cryptoutil.WrapAES128Key(ctx, appSKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

var errAppKeyNotExposed = errors.DefineFailedPrecondition("app_key_not_exposed", "AppKey not exposed")

func (s applicationCryptoServiceServer) GetAppKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Application == nil {
		return nil, errNoApplicationService.WithAttributes("id", req.ProvisionerID)
	}
	if !provisioner.ExposeRootKeys {
		return nil, errAppKeyNotExposed.New()
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	appKey, err := provisioner.Application.GetAppKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	env, err := cryptoutil.WrapAES128Key(ctx, *appKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
