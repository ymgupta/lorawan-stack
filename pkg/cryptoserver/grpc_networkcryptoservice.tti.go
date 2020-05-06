// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type networkCryptoServiceServer struct {
	Provisioners Provisioners
	KeyVault     crypto.KeyVault
}

var errNoNetworkService = errors.DefineFailedPrecondition("no_network_service", "no network service provided by provisioner `{id}`")

func (s networkCryptoServiceServer) JoinRequestMIC(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	mic, err := provisioner.Network.JoinRequestMIC(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s networkCryptoServiceServer) JoinAcceptMIC(ctx context.Context, req *ttnpb.JoinAcceptMICRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	mic, err := provisioner.Network.JoinAcceptMIC(ctx, dev, req.LoRaWANVersion, byte(req.JoinRequestType), req.DevNonce, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: mic[:],
	}, nil
}

func (s networkCryptoServiceServer) EncryptJoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	data, err := provisioner.Network.EncryptJoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s networkCryptoServiceServer) EncryptRejoinAccept(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	data, err := provisioner.Network.EncryptRejoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
	if err != nil {
		return nil, err
	}
	return &ttnpb.CryptoServicePayloadResponse{
		Payload: data,
	}, nil
}

func (s networkCryptoServiceServer) DeriveNwkSKeys(ctx context.Context, req *ttnpb.DeriveSessionKeysRequest) (*ttnpb.NwkSKeysResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	nwkSKeys, err := provisioner.Network.DeriveNwkSKeys(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	res := &ttnpb.NwkSKeysResponse{}
	res.FNwkSIntKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.FNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.SNwkSIntKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.SNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.NwkSEncKey, err = cryptoutil.WrapAES128Key(ctx, nwkSKeys.NwkSEncKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

var errNwkKeyNotExposed = errors.DefineFailedPrecondition("nwk_key_not_exposed", "NwkKey not exposed")

func (s networkCryptoServiceServer) GetNwkKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	provisioner, err := s.Provisioners.Get(req.ProvisionerID)
	if err != nil {
		return nil, err
	}
	if provisioner.Network == nil {
		return nil, errNoNetworkService.WithAttributes("id", req.ProvisionerID)
	}
	if !provisioner.ExposeRootKeys {
		return nil, errNwkKeyNotExposed.New()
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	nwkKey, err := provisioner.Network.GetNwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	env, err := cryptoutil.WrapAES128Key(ctx, *nwkKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
