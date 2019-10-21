// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/stripe"
)

// Config is the configuration for the TenantBillingServer.
type Config struct {
	Stripe stripe.Config `name:"stripe" description:"Stripe backend configuration"`

	TenantAdminKey string `name:"tenant-admin-key" description:"Tenant administration authentication key"`

	ReporterKeys        []string `name:"reporter-keys" description:"Keys that can be used for billing reporting"`
	decodedReporterKeys [][]byte
}

func (c *Config) decodeKeys(ctx context.Context) error {
	for i, key := range c.ReporterKeys {
		decodedKey, err := hex.DecodeString(key)
		if err != nil {
			return errInvalidBillingReporterKey.WithCause(err)
		}
		switch len(decodedKey) {
		case 16, 24, 32:
		default:
			return errInvalidBillingReporterKey.WithCause(fmt.Errorf("invalid length for key %d: must be 16, 24 or 32 bytes", i))
		}
		c.decodedReporterKeys = append(c.decodedReporterKeys, decodedKey)
	}
	if c.decodedReporterKeys == nil {
		c.decodedReporterKeys = [][]byte{random.Bytes(32)}
		log.FromContext(ctx).WithField("key", hex.EncodeToString(c.decodedReporterKeys[0])).Warn("No billing admin key configured, generated a random one")
	}
	return nil
}
