// Copyright © 2020 The Things Industries B.V.

package formatters

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	stdjson "encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type awsVioletDevEUI types.EUI64

// MarshalText marshals the DevEUI as text in the form
// "00-00-00-00-00-00-00-00".
func (eui awsVioletDevEUI) MarshalText() ([]byte, error) {
	b, err := types.EUI64(eui).MarshalText()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, 16+7)
	for i := 0; i < 8; i++ {
		out = append(out, b[i*2:(i+1)*2]...)
		if i != 7 {
			out = append(out, '-')
		}
	}
	return out, nil
}

type awsVioletTimestamp time.Time

// MarshalJSON marshals the timestamp as JSON number in the form unix.nanos.
func (t awsVioletTimestamp) MarshalJSON() ([]byte, error) {
	gt := time.Time(t)
	return []byte(fmt.Sprintf("%d.%d", gt.Unix(), gt.Nanosecond()/1000)), nil
}

type awsVioletGatewayEUI types.EUI64

// MarshalJSON marshals the EUI as a JSON number.
func (eui awsVioletGatewayEUI) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(types.EUI64(eui).MarshalNumber(), 10)), nil
}

type awsVioletDevAddr types.DevAddr

// MarshalJSON marshals the DevAddr as a JSON number.
func (addr awsVioletDevAddr) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(types.DevAddr(addr).MarshalNumber()), 10)), nil
}

type awsVioletKeyEnvelope ttnpb.KeyEnvelope

// MarshalText marshals the KeyEnvelope as text in the form
//     "plain:" + HEX(Key)
// or
//     "aes:" + BASE64(0x00 | KEKID | WrappedKey)
// where KEKID is hex_decode(KEKLabel)
func (key awsVioletKeyEnvelope) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	if key.Key != nil {
		buf.WriteString("plain:")
		enc := hex.NewEncoder(&buf)
		enc.Write(key.Key[:])
	} else {
		buf.WriteString("aes:")
		enc := base64.NewEncoder(base64.StdEncoding, &buf)
		enc.Write([]byte{0x00})
		kekID, err := hex.DecodeString(key.KEKLabel)
		if err != nil {
			return nil, err
		}
		enc.Write(kekID)
		enc.Write(key.EncryptedKey)
		enc.Close()
	}
	return buf.Bytes(), nil
}

type awsVioletPayload []byte

// MarshalText marshals the payload as hexadecimal text.
func (pld awsVioletPayload) MarshalText() ([]byte, error) {
	b := make([]byte, hex.EncodedLen(len(pld)))
	hex.Encode(b, pld)
	return b, nil
}

// UpMeta represents uplink gateway info.
type UpMeta struct {
	SNR      float32             `json:"snr"`
	RSSI     float32             `json:"rssi"`
	MuxID    uint8               `json:"muxid"`
	XTime    uint32              `json:"xtime"`
	DoorID   uint8               `json:"doorid"`
	RxTime   *awsVioletTimestamp `json:"rxtime,omitempty"`
	ArrTime  *awsVioletTimestamp `json:"ArrTime,omitempty"`
	RxDelay  uint8               `json:"RxDelay"`
	Rx1DrOff uint8               `json:"RX1DRoff"`
	RegionID uint64              `json:"regionid"`
	RouterID awsVioletGatewayEUI `json:"routerid"`
}

type UpDF struct {
	DR         uint8                 `json:"DR"`
	Freq       uint64                `json:"Freq"`
	FPort      uint8                 `json:"FPort"`
	DevEUI     awsVioletDevEUI       `json:"DevEUI"`
	FCntUp     uint32                `json:"FCntUp"`
	SessID     uint32                `json:"SessID"`
	DClass     string                `json:"dClass"`
	Region     string                `json:"region"`
	ArrTime    *awsVioletTimestamp   `json:"ArrTime,omitempty"`
	UpInfo     []UpMeta              `json:"upinfo,omitempty"`
	DevAddr    awsVioletDevAddr      `json:"DevAddr"`
	Confirm    bool                  `json:"confirm"`
	MsgType    string                `json:"msgtype"`
	Ciphered   bool                  `json:"ciphered"`
	RegionID   uint32                `json:"regionid"`
	AFCntDown  uint32                `json:"AFCntDown"`
	AppSKeyEnv *awsVioletKeyEnvelope `json:"AppSKeyEnv,omitempty"`
	FRMPayload awsVioletPayload      `json:"FRMPayload"`
	UpID       uint64                `json:"upid"`
}

type awsviolet struct {
}

func (awsviolet) FromUp(msg *ttnpb.ApplicationUp) ([]byte, error) {
	var pld interface{}

	switch oneof := msg.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		up := oneof.UplinkMessage
		updf := UpDF{
			DR:         uint8(up.Settings.DataRateIndex),
			Freq:       up.Settings.Frequency,
			FPort:      uint8(up.FPort),
			FCntUp:     up.FCnt,
			DClass:     "UNKNOWN",         // Seems to be unused.
			Region:     "UNKNOWN",         // Seems to be unused.
			Confirm:    false,             // Seems to be unused.
			Ciphered:   up.AppSKey != nil, // FRMPayload is still encrypted.
			RegionID:   1000,              // Seems to be unused.
			FRMPayload: awsVioletPayload(up.FRMPayload),
			UpID:       0, // Seems to be unused.
		}
		if msg.DevEUI != nil {
			updf.DevEUI = awsVioletDevEUI(*msg.DevEUI)
		}
		if len(up.SessionKeyID) > 0 {
			updf.SessID = crc32.ChecksumIEEE(up.SessionKeyID)
		}
		if msg.DevAddr != nil {
			updf.DevAddr = awsVioletDevAddr(*msg.DevAddr)
		}
		if up.Settings.Time != nil {
			tmst := awsVioletTimestamp(*up.Settings.Time)
			updf.ArrTime = &tmst
		} else {
			tmst := awsVioletTimestamp(up.ReceivedAt)
			updf.ArrTime = &tmst
		}
		if up.AppSKey != nil {
			env := awsVioletKeyEnvelope(*up.AppSKey)
			updf.AppSKeyEnv = &env
			updf.AFCntDown = up.LastAFCntDown
		}
		if len(up.RxMetadata) > 0 {
			updf.MsgType = "upinfo"
			for _, meta := range up.RxMetadata {
				upinfo := UpMeta{
					SNR:      meta.SNR,
					RSSI:     meta.RSSI,
					MuxID:    0, // Seems to be unused.
					XTime:    meta.Timestamp,
					DoorID:   0,    // Seems to be unused.
					RxDelay:  0,    // This is network layer.
					Rx1DrOff: 0,    // This is network layer.
					RegionID: 1000, // Seems to be unused.
				}
				if meta.Time != nil {
					tmst := awsVioletTimestamp(*meta.Time)
					upinfo.RxTime = &tmst
					upinfo.ArrTime = &tmst
				}
				if meta.GatewayIdentifiers.EUI != nil {
					upinfo.RouterID = awsVioletGatewayEUI(*meta.EUI)
				}
				updf.UpInfo = append(updf.UpInfo, upinfo)
			}
		} else {
			updf.MsgType = "updf"
		}
		pld = updf
	}

	return stdjson.Marshal(pld)
}

func (awsviolet) ToDownlinks(data []byte) (*ttnpb.ApplicationDownlinks, error) {
	res := &ttnpb.ApplicationDownlinks{}
	if err := jsonpb.TTN().Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (awsviolet) ToDownlinkQueueRequest(data []byte) (*ttnpb.DownlinkQueueRequest, error) {
	res := &ttnpb.DownlinkQueueRequest{}
	if err := jsonpb.TTN().Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// AWSViolet is a formatter for AWS Violet that uses JSON marshaling.
var AWSViolet Formatter = &awsviolet{}
