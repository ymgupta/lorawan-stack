// Copyright © 2019 The Things Industries B.V.

//+build tti

package unique

import (
	"context"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errUniqueIdentifier = errors.DefineInvalidArgument("unique_identifier", "invalid unique identifier `{uid}`")
	errFormat           = errors.DefineInvalidArgument("format", "invalid format in value `{value}`")
	errMissingTenantID  = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")
	errStartsWithADot   = errors.DefineInvalidArgument("starts_with_a_dot", "starts with a dot")
	errEndsWithADot     = errors.DefineInvalidArgument("ends_with_a_dot", "ends with a dot")
	errEmpty            = errors.DefineInvalidArgument("empty", "empty")
)

// ID returns the unique identifier of the specified identifiers.
// This function panics if the resulting identifier is invalid.
// The reason for panicking is that taking the unique identifier of a nil or
// zero value may result in unexpected and potentially harmful behavior.
func ID(ctx context.Context, id ttnpb.Identifiers) (res string) {
	if idStringer, ok := id.(interface{ IDString() string }); ok {
		res = idStringer.IDString()
	} else {
		eids := id.CombinedIdentifiers().EntityIdentifiers
		if len(eids) != 1 {
			panic(fmt.Errorf("failed to determine unique ID: invalid number of identifiers for unique ID"))
		}
		res = eids[0].IDString()
	}
	switch {
	case res == "":
		panic(errUniqueIdentifier.WithAttributes("uid", res).WithCause(errEmpty))
	case strings.HasPrefix(res, "."):
		panic(errUniqueIdentifier.WithAttributes("uid", res).WithCause(errStartsWithADot))
	case strings.HasSuffix(res, "."):
		panic(errUniqueIdentifier.WithAttributes("uid", res).WithCause(errEndsWithADot))
	}

	if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
		return fmt.Sprintf("%s@%s", res, tenantID)
	}
	panic(errMissingTenantID)
}

func parse(uid string) (tenantID, id string, err error) {
	if sepIdx := strings.Index(uid, "@"); sepIdx != -1 {
		return uid[sepIdx+1:], uid[:sepIdx], nil
	}
	return "", "", errMissingTenantID.New()
}

// ToTenantID returns the tenant identifier of the specified unique ID.
func ToTenantID(uid string) (id ttipb.TenantIdentifiers, err error) {
	id.TenantID, _, err = parse(uid)
	return
}

// WithContext returns the given context with tenant identifier.
func WithContext(ctx context.Context, uid string) (context.Context, error) {
	tenantID, err := ToTenantID(uid)
	if err != nil {
		return nil, err
	}
	return log.NewContextWithField(tenant.NewContext(ctx, tenantID), "tenant_id", tenantID.TenantID), nil
}

// ToApplicationID returns the application identifier of the specified unique ID.
func ToApplicationID(uid string) (id ttnpb.ApplicationIdentifiers, err error) {
	_, id.ApplicationID, err = parse(uid)
	if err != nil {
		return
	}
	if err := id.ValidateFields("application_id"); err != nil {
		return ttnpb.ApplicationIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToClientID returns the client identifier of the specified unique ID.
func ToClientID(uid string) (id ttnpb.ClientIdentifiers, err error) {
	_, id.ClientID, err = parse(uid)
	if err != nil {
		return
	}
	if err := id.ValidateFields("client_id"); err != nil {
		return ttnpb.ClientIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToDeviceID returns the end device identifier of the specified unique ID.
func ToDeviceID(uid string) (id ttnpb.EndDeviceIdentifiers, err error) {
	_, uid, err = parse(uid)
	if err != nil {
		return
	}
	sepIdx := strings.Index(uid, ".")
	if sepIdx == -1 {
		return ttnpb.EndDeviceIdentifiers{}, errFormat.WithAttributes("uid", uid)
	}
	id.ApplicationIdentifiers.ApplicationID = uid[:sepIdx]
	id.DeviceID = uid[sepIdx+1:]
	if err := id.ValidateFields("device_id", "application_ids"); err != nil {
		return ttnpb.EndDeviceIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToGatewayID returns the gateway identifier of the specified unique ID.
func ToGatewayID(uid string) (id ttnpb.GatewayIdentifiers, err error) {
	_, id.GatewayID, err = parse(uid)
	if err != nil {
		return
	}
	if err := id.ValidateFields("gateway_id"); err != nil {
		return ttnpb.GatewayIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToOrganizationID returns the organization identifier of the specified unique ID.
func ToOrganizationID(uid string) (id ttnpb.OrganizationIdentifiers, err error) {
	_, id.OrganizationID, err = parse(uid)
	if err != nil {
		return
	}
	if err := id.ValidateFields("organization_id"); err != nil {
		return ttnpb.OrganizationIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToUserID returns the user identifier of the specified unique ID.
func ToUserID(uid string) (id ttnpb.UserIdentifiers, err error) {
	_, id.UserID, err = parse(uid)
	if err != nil {
		return
	}
	if err := id.ValidateFields("user_id"); err != nil {
		return ttnpb.UserIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}
