// Copyright © 2020 The Things Industries B.V.

package tenantbillingserver

import (
	"context"
	"regexp"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/stripe"
)

// Config is the configuration for the TenantBillingServer.
type Config struct {
	Stripe stripe.Config `name:"stripe" description:"Stripe backend configuration"`

	TenantAdminKey string `name:"tenant-admin-key" description:"Tenant administration authentication key"`

	PullInterval           time.Duration `name:"pull-interval" description:"How frequently to pull the metering data"`
	ReporterAddressRegexps []string      `name:"reporter-address-regexps" description:"Regular expressions of addresses that can report metering data"`
	reporterAddressRegexps []*regexp.Regexp
}

var (
	errInvalidReporterAddressRegexp = errors.DefineInvalidArgument("invalid_reporter_address_regexp", "invalid reporter address regular expression")
)

func (c *Config) compileRegex(ctx context.Context) error {
	for _, addr := range c.ReporterAddressRegexps {
		r, err := regexp.Compile(addr)
		if err != nil {
			return errInvalidReporterAddressRegexp.WithCause(err)
		}
		c.reporterAddressRegexps = append(c.reporterAddressRegexps, r)
	}
	return nil
}
