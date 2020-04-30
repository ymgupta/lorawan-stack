// Copyright © 2019 The Things Industries B.V.

package license

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type licenseContextKeyType struct{}

var licenseContextKey licenseContextKeyType

var now = time.Now()

var defaultLicense = ttipb.License{
	LicenseIdentifiers:      ttipb.LicenseIdentifiers{LicenseID: "testing"},
	CreatedAt:               now,
	ValidFrom:               now,
	ValidUntil:              now.Add(10 * time.Minute),
	ComponentAddressRegexps: []string{"localhost"},
	MultiTenancy:            true,
}

// NewContextWithLicense returns a context derived from parent that contains license.
func NewContextWithLicense(parent context.Context, license ttipb.License) context.Context {
	return context.WithValue(parent, licenseContextKey, license)
}

// FromContext returns the License from the context if present. Otherwise it returns default License.
func FromContext(ctx context.Context) ttipb.License {
	if license, ok := ctx.Value(licenseContextKey).(ttipb.License); ok {
		if license.Metering != nil {
			license = globalMetering.Apply(license)
		}
		return license
	}
	return defaultLicense
}

var errComponentNotLicensed = errors.DefineFailedPrecondition("component_not_licensed", "the `{component}` component is not included in this license", "licensed")

// RequireComponent requires components to be included in the license.
func RequireComponent(ctx context.Context, components ...ttnpb.ClusterRole) error {
	license := FromContext(ctx)
	if err := CheckValidity(&license); err != nil {
		return err
	}
	if len(license.Components) == 0 {
		return nil
	}
nextComponent:
	for _, component := range components {
		for _, licensed := range license.Components {
			if component == licensed {
				continue nextComponent
			}
		}
		return errComponentNotLicensed.WithAttributes("component", strings.Title(strings.Replace(component.String(), "_", " ", -1)), "licensed", license.Components)
	}
	return nil
}

var errMultiTenancyNotLicensed = errors.DefineFailedPrecondition("multi_tenancy_not_licensed", "multi-tenancy is not included in this license")

// RequireMultiTenancy requires multi-tenancy to be included in the license.
func RequireMultiTenancy(ctx context.Context) error {
	license := FromContext(ctx)
	if err := CheckValidity(&license); err != nil {
		return err
	}
	if !license.MultiTenancy {
		return errMultiTenancyNotLicensed.New()
	}
	return nil
}

var componentAddressRegexps sync.Map

type componentAddressRegexp struct {
	*regexp.Regexp
	wait chan struct{}
	err  error
}

func getComponentAddressRegexp(s string) (*regexp.Regexp, error) {
	reI, loaded := componentAddressRegexps.LoadOrStore(s, &componentAddressRegexp{wait: make(chan struct{})})
	re := reI.(*componentAddressRegexp)
	if !loaded {
		re.Regexp, re.err = regexp.Compile(s)
		close(re.wait)
	}
	return re.Regexp, re.err
}

var errComponentAddressNotLicensed = errors.DefineFailedPrecondition("component_address_not_licensed", "component address `{address}` is not included in this license", "licensed")

// RequireComponentAddress requires the given address to be included in the license.
func RequireComponentAddress(ctx context.Context, addr string) error {
	if addr == "" {
		return nil
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		url, err := url.Parse(addr)
		if err != nil {
			return err
		}
		addr = url.Hostname()
	}
	license := FromContext(ctx)
	if err := CheckValidity(&license); err != nil {
		return err
	}
	if regexps := license.GetComponentAddressRegexps(); len(regexps) > 0 {
		for _, regexp := range regexps {
			regexp, err := getComponentAddressRegexp(regexp)
			if err != nil {
				continue
			}
			if regexp.MatchString(addr) {
				return nil
			}
		}
		return errComponentAddressNotLicensed.WithAttributes("address", addr, "licensed", regexps)
	}
	return nil
}

var errDevAddrPrefixNotLicensed = errors.DefineFailedPrecondition("dev_addr_prefix_not_licensed", "DevAddr prefix `{prefix}` is not included in this license", "licensed")

// RequireDevAddrPrefix requires the given DevAddrPrefix to be included in the license.
func RequireDevAddrPrefix(ctx context.Context, prefix types.DevAddrPrefix) error {
	license := FromContext(ctx)
	if err := CheckValidity(&license); err != nil {
		return err
	}
	if licensedPrefixes := license.DevAddrPrefixes; len(licensedPrefixes) > 0 {
		minAddr := types.DevAddr{0x00, 0x00, 0x00, 0x00}.WithPrefix(prefix)
		maxAddr := types.DevAddr{0xff, 0xff, 0xff, 0xff}.WithPrefix(prefix)
		for _, licensedPrefix := range licensedPrefixes {
			if licensedPrefix.Matches(minAddr) && licensedPrefix.Matches(maxAddr) {
				return nil
			}
		}
		return errDevAddrPrefixNotLicensed.WithAttributes("prefix", prefix.String(), "licensed", licensedPrefixes)
	}
	return nil
}

var errJoinEUIPrefixNotLicensed = errors.DefineFailedPrecondition("join_eui_prefix_not_licensed", "JoinEUI prefix `{prefix}` is not included in this license", "licensed")

// RequireJoinEUIPrefix requires the given JoinEUI prefix to be included in the license.
func RequireJoinEUIPrefix(ctx context.Context, prefix types.EUI64Prefix) error {
	license := FromContext(ctx)
	if err := CheckValidity(&license); err != nil {
		return err
	}
	if licensedPrefixes := license.JoinEUIPrefixes; len(licensedPrefixes) > 0 {
		minAddr := types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}.WithPrefix(prefix)
		maxAddr := types.EUI64{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.WithPrefix(prefix)
		for _, licensedPrefix := range licensedPrefixes {
			if licensedPrefix.Matches(minAddr) && licensedPrefix.Matches(maxAddr) {
				return nil
			}
		}
		return errJoinEUIPrefixNotLicensed.WithAttributes("prefix", prefix.String(), "licensed", licensedPrefixes)
	}
	return nil
}
