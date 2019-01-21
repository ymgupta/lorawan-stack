// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"
	"encoding/hex"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errProvisioner  = errors.DefineCorruption("provisioner", "invalid provisioner `{id}`")
	errPartNumber   = errors.DefineCorruption("part_number", "invalid part number `{part_number}`")
	errSerialNumber = errors.DefineCorruption("serial_number", "invalid serial number `{serial_number}`")
)

func (m *impl) getRootKeys(dev *ttnpb.EndDevice) (nwkKey, appKey types.AES128Key, err error) {
	if dev.ProvisionerID != "microchip" || dev.ProvisioningData == nil {
		return types.AES128Key{}, types.AES128Key{}, errProvisioner.WithAttributes("id", dev.ProvisionerID)
	}
	part := dev.ProvisioningData.Fields["partNumber"].GetStringValue()

	m.parentKeysMu.RLock()
	key, ok := m.parentKeys[part]
	m.parentKeysMu.RUnlock()
	if !ok {
		return types.AES128Key{}, types.AES128Key{}, errPartNumber.WithAttributes("part_number", part)
	}

	snHex := dev.ProvisioningData.Fields["uniqueId"].GetStringValue()
	snBuf, err := hex.DecodeString(snHex)
	if err != nil {
		return types.AES128Key{}, types.AES128Key{}, errSerialNumber.WithAttributes("serial_number", snHex)
	}
	var sn SerialNumber
	copy(sn[:], snBuf)

	nwkKey, appKey = DiversifiedRootKeys(key, sn)
	return
}

func (m *impl) JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([4]byte, error) {
	return [4]byte{}, nil
}

func (m *impl) JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error) {
	return [4]byte{}, nil
}

func (m *impl) EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	return nil, nil
}

func (m *impl) EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	return nil, nil
}

func (m *impl) DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (cryptoservices.NwkSKeys, error) {
	return cryptoservices.NwkSKeys{}, nil
}

func (m *impl) NwkKey(ctx context.Context, dev *ttnpb.EndDevice) (types.AES128Key, error) {
	nwkKey, _, err := m.getRootKeys(dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	return nwkKey, err
}

func (m *impl) DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	return types.AES128Key{}, nil
}

func (m *impl) AppKey(ctx context.Context, dev *ttnpb.EndDevice) (types.AES128Key, error) {
	_, appKey, err := m.getRootKeys(dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	return appKey, err
}
