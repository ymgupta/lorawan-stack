// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"
	"encoding/hex"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errProvisioner  = errors.DefineCorruption("provisioner", "invalid provisioner `{id}`")
	errPartNumber   = errors.DefineCorruption("part_number", "invalid part number `{part_number}`")
	errSerialNumber = errors.DefineCorruption("serial_number", "invalid serial number `{serial_number}`")
	errNoJoinEUI    = errors.DefineInvalidArgument("no_join_eui", "no JoinEUI specified")
	errNoDevEUI     = errors.DefineInvalidArgument("no_dev_eui", "no DevEUI specified")
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

func (m *impl) getNwkKey(dev *ttnpb.EndDevice, version ttnpb.MACVersion) (types.AES128Key, error) {
	nwkKey, appKey, err := m.getRootKeys(dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		return nwkKey, nil
	default:
		return appKey, nil
	}
}

func (m *impl) getAppKey(dev *ttnpb.EndDevice, version ttnpb.MACVersion) (types.AES128Key, error) {
	_, appKey, err := m.getRootKeys(dev)
	if err != nil {
		return types.AES128Key{}, err
	}
	return appKey, nil
}

func (m *impl) JoinRequestMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([4]byte, error) {
	key, err := m.getNwkKey(dev, version)
	if err != nil {
		return [4]byte{}, err
	}
	return crypto.ComputeJoinRequestMIC(key, payload)
}

func (m *impl) JoinAcceptMIC(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, joinReqType byte, dn types.DevNonce, payload []byte) ([4]byte, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return [4]byte{}, errNoJoinEUI.New()
	}
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return [4]byte{}, errNoDevEUI.New()
	}
	key, err := m.getNwkKey(dev, version)
	if err != nil {
		return [4]byte{}, err
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		jsIntKey := crypto.DeriveJSIntKey(key, *dev.DevEUI)
		return crypto.ComputeJoinAcceptMIC(jsIntKey, joinReqType, *dev.JoinEUI, dn, payload)
	default:
		return crypto.ComputeLegacyJoinAcceptMIC(key, payload)
	}
}

func (m *impl) EncryptJoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	key, err := m.getNwkKey(dev, version)
	if err != nil {
		return nil, err
	}
	return crypto.EncryptJoinAccept(key, payload)
}

var errMACVersion = errors.DefineCorruption("mac_version", "invalid MAC version")

func (m *impl) EncryptRejoinAccept(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, payload []byte) ([]byte, error) {
	if dev.DevEUI == nil || dev.DevEUI.IsZero() {
		return payload, errNoDevEUI.New()
	}
	if version.Compare(ttnpb.MAC_V1_1) < 0 {
		return nil, errMACVersion.New()
	}
	nwkKey, err := m.getNwkKey(dev, version)
	if err != nil {
		return nil, err
	}
	jsEncKey := crypto.DeriveJSEncKey(nwkKey, *dev.DevEUI)
	return crypto.EncryptJoinAccept(jsEncKey, payload)
}

func (m *impl) DeriveNwkSKeys(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (cryptoservices.NwkSKeys, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return cryptoservices.NwkSKeys{}, errNoJoinEUI.New()
	}
	key, err := m.getNwkKey(dev, version)
	if err != nil {
		return cryptoservices.NwkSKeys{}, err
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		return cryptoservices.NwkSKeys{
			FNwkSIntKey: crypto.DeriveFNwkSIntKey(key, jn, *dev.JoinEUI, dn),
			SNwkSIntKey: crypto.DeriveSNwkSIntKey(key, jn, *dev.JoinEUI, dn),
			NwkSEncKey:  crypto.DeriveNwkSEncKey(key, jn, *dev.JoinEUI, dn),
		}, nil
	default:
		return cryptoservices.NwkSKeys{
			FNwkSIntKey: crypto.DeriveLegacyNwkSKey(key, jn, nid, dn),
		}, nil
	}
}

func (m *impl) GetNwkKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	nwkKey, _, err := m.getRootKeys(dev)
	if err != nil {
		return nil, err
	}
	return &nwkKey, err
}

func (m *impl) DeriveAppSKey(ctx context.Context, dev *ttnpb.EndDevice, version ttnpb.MACVersion, jn types.JoinNonce, dn types.DevNonce, nid types.NetID) (types.AES128Key, error) {
	if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
		return types.AES128Key{}, errNoJoinEUI.New()
	}
	appKey, err := m.getAppKey(dev, version)
	if err != nil {
		return types.AES128Key{}, err
	}
	switch {
	case version.Compare(ttnpb.MAC_V1_1) >= 0:
		return crypto.DeriveAppSKey(appKey, jn, *dev.JoinEUI, dn), nil
	default:
		return crypto.DeriveLegacyAppSKey(appKey, jn, nid, dn), nil
	}
}

func (m *impl) GetAppKey(ctx context.Context, dev *ttnpb.EndDevice) (*types.AES128Key, error) {
	_, appKey, err := m.getRootKeys(dev)
	if err != nil {
		return nil, err
	}
	return &appKey, err
}
