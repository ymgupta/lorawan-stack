// Copyright © 2019 The Things Industries B.V.

package identityserver

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
)

// TenancyConfig is the configuration for tenancy in the Identity Server.
type TenancyConfig struct {
	AdminKeys []string `name:"admin-keys" description:"Keys that can be used for tenant administration"`

	decodedAdminKeys [][]byte
}

func (c *TenancyConfig) decodeAdminKeys(ctx context.Context) error {
	for i, key := range c.AdminKeys {
		decodedKey, err := hex.DecodeString(key)
		if err != nil {
			return errInvalidTenantAdminKey.WithCause(err)
		}
		switch len(decodedKey) {
		case 16, 24, 32:
		default:
			return errInvalidTenantAdminKey.WithCause(fmt.Errorf("invalid length for key %d: must be 16, 24 or 32 bytes", i))
		}
		c.decodedAdminKeys = append(c.decodedAdminKeys, decodedKey)
	}
	if c.decodedAdminKeys == nil {
		c.decodedAdminKeys = [][]byte{random.Bytes(32)}
		log.FromContext(ctx).WithField("key", hex.EncodeToString(c.decodedAdminKeys[0])).Warn("No tenant admin key configured, generated a random one")
	}
	return nil
}
