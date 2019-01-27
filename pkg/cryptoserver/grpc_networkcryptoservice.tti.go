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

type networkCryptoServiceServer struct {
	Network        cryptoservices.Network
	KeyVault       crypto.KeyVault
	ExposeRootKeys bool
}

func (s networkCryptoServiceServer) JoinRequestMIC(ctx context.Context, req *ttnpb.CryptoServicePayloadRequest) (*ttnpb.CryptoServicePayloadResponse, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	mic, err := s.Network.JoinRequestMIC(ctx, dev, req.LoRaWANVersion, req.Payload)
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
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	mic, err := s.Network.JoinAcceptMIC(ctx, dev, req.LoRaWANVersion, byte(req.JoinRequestType), req.DevNonce, req.Payload)
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
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	data, err := s.Network.EncryptJoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
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
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	data, err := s.Network.EncryptRejoinAccept(ctx, dev, req.LoRaWANVersion, req.Payload)
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
	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: req.EndDeviceIdentifiers,
		ProvisionerID:        req.ProvisionerID,
		ProvisioningData:     req.ProvisioningData,
	}
	nwkSKeys, err := s.Network.DeriveNwkSKeys(ctx, dev, req.LoRaWANVersion, req.JoinNonce, req.DevNonce, req.NetID)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	res := &ttnpb.NwkSKeysResponse{}
	res.FNwkSIntKey, err = cryptoutil.WrapAES128Key(nwkSKeys.FNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.SNwkSIntKey, err = cryptoutil.WrapAES128Key(nwkSKeys.SNwkSIntKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	res.NwkSEncKey, err = cryptoutil.WrapAES128Key(nwkSKeys.NwkSEncKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s networkCryptoServiceServer) GetNwkKey(ctx context.Context, req *ttnpb.GetRootKeysRequest) (*ttnpb.KeyEnvelope, error) {
	if err := cluster.Authorized(ctx); err != nil {
		return nil, err
	}
	if s.Network == nil {
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
	nwkKey, err := s.Network.GetNwkKey(ctx, dev)
	if err != nil {
		return nil, err
	}
	// TODO: Encrypt root keys (https://github.com/thethingsindustries/lorawan-stack/issues/1562)
	env, err := cryptoutil.WrapAES128Key(nwkKey, "", s.KeyVault)
	if err != nil {
		return nil, err
	}
	return &env, nil
}
