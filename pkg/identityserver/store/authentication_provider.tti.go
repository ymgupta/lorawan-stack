// Copyright © 2020 The Things Industries B.V.

package store

import (
	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm/dialects/postgres"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// External user model.
type AuthenticationProvider struct {
	Model
	SoftDelete

	TenantID string `gorm:"unique_index:provider_id_index;type:VARCHAR(36)"`

	ProviderID         string `gorm:"unique_index:provider_id_index;type:VARCHAR(36);not null"`
	Name               string `gorm:"type:VARCHAR"`
	AllowRegistrations bool

	Configuration postgres.Jsonb `gorm:"not null"`
}

func (AuthenticationProvider) _isMultiTenant() {}

func init() {
	registerModel(&AuthenticationProvider{})
}

// functions to set fields from the authentication provider model into the authentication provider proto.
var authenticationProviderPBSetters = map[string]func(*ttipb.AuthenticationProvider, *AuthenticationProvider){
	nameField: func(pb *ttipb.AuthenticationProvider, ap *AuthenticationProvider) { pb.Name = ap.Name },
	allowRegistrationsField: func(pb *ttipb.AuthenticationProvider, ap *AuthenticationProvider) {
		pb.AllowRegistrations = ap.AllowRegistrations
	},
}

// functions to set fields from the authentication provider proto into the authentication provider model.
var authenticationProviderModelSetters = map[string]func(*AuthenticationProvider, *ttipb.AuthenticationProvider){
	nameField: func(ap *AuthenticationProvider, pb *ttipb.AuthenticationProvider) { ap.Name = pb.Name },
	allowRegistrationsField: func(ap *AuthenticationProvider, pb *ttipb.AuthenticationProvider) {
		ap.AllowRegistrations = pb.AllowRegistrations
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultAuthenticationProviderFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(authenticationProviderPBSetters))
	for path := range authenticationProviderPBSetters {
		paths = append(paths, path)
	}
	paths = append(paths, "configuration")
	defaultAuthenticationProviderFieldMask.Paths = paths
}

// fieldmask path to column name in authentication providers table.
var authenticationProviderColumnNames = map[string][]string{
	nameField:               {nameField},
	allowRegistrationsField: {allowRegistrationsField},
	configurationField:      {configurationField},
}

func (ap AuthenticationProvider) toPB(pb *ttipb.AuthenticationProvider, fieldMask *types.FieldMask) error {
	pb.AuthenticationProviderIdentifiers.ProviderID = ap.ProviderID
	pb.CreatedAt = cleanTime(ap.CreatedAt)
	pb.UpdatedAt = cleanTime(ap.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultAuthenticationProviderFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := authenticationProviderPBSetters[path]; ok {
			setter(pb, &ap)
		}
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), configurationField) {
		if len(ap.Configuration.RawMessage) > 0 {
			var configuration ttipb.AuthenticationProvider_Configuration
			if err := jsonpb.TTN().Unmarshal(ap.Configuration.RawMessage, &configuration); err != nil {
				return err
			}
			pb.Configuration = &configuration
		} else {
			pb.Configuration = nil
		}
		tmp := &ttipb.AuthenticationProvider{}
		if err := tmp.SetFields(pb, fieldMask.Paths...); err != nil {
			return err
		}
		pb.Configuration = tmp.Configuration
	}
	return nil
}

func (ap *AuthenticationProvider) fromPB(pb *ttipb.AuthenticationProvider, fieldMask *types.FieldMask) (columns []string, err error) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultAuthenticationProviderFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := authenticationProviderModelSetters[path]; ok {
			setter(ap, pb)
			if columnNames, ok := authenticationProviderColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			} else {
				columns = append(columns, path)
			}
			continue
		}
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), configurationField) {
		var configuration ttipb.AuthenticationProvider_Configuration
		if len(ap.Configuration.RawMessage) > 0 {
			if err = jsonpb.TTN().Unmarshal(ap.Configuration.RawMessage, &configuration); err != nil {
				return nil, err
			}
		}
		tmp := &ttipb.AuthenticationProvider{Configuration: &configuration}
		if err := tmp.SetFields(pb, fieldMask.Paths...); err != nil {
			return nil, err
		}
		if tmp.Configuration == nil {
			ap.Configuration.RawMessage = nil
		} else if ap.Configuration.RawMessage, err = jsonpb.TTN().Marshal(tmp.Configuration); err != nil {
			return nil, err
		}
		columns = append(columns, configurationField)
	}
	return
}
