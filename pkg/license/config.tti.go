// Copyright © 2019 The Things Industries B.V.

package license

import (
	"encoding/base64"
	"io/ioutil"

	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

// Config represents the license configuration.
type Config struct {
	File string `name:"file" description:"Location of the license file"`
	Key  string `name:"key" description:"Contents of the license key"`
}

// Read reads and validates the license using the given config and returns it.
func Read(config Config) (*ttipb.License, error) {
	var (
		licenseKeyBytes []byte
		err             error
	)
	switch {
	case config.Key != "":
		licenseKeyBytes, err = base64.StdEncoding.DecodeString(config.Key)
	case config.File != "":
		licenseKeyBytes, err = ioutil.ReadFile(config.File)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var licenseKey ttipb.LicenseKey
	if err = licenseKey.Unmarshal(licenseKeyBytes); err != nil {
		return nil, errInvalidLicense.WithCause(err)
	}
	license, err := VerifyKey(&licenseKey)
	if err != nil {
		return nil, err
	}
	return &license, nil
}
