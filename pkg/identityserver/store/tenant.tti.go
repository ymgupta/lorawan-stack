// Copyright © 2019 The Things Industries B.V.

package store

import (
	"encoding/json"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm/dialects/postgres"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Tenant model.
type Tenant struct {
	Model
	SoftDelete

	// BEGIN common fields
	TenantID    string      `gorm:"unique_index:tenant_id_index;type:VARCHAR(36);not null"`
	Name        string      `gorm:"type:VARCHAR"`
	Description string      `gorm:"type:TEXT"`
	Attributes  []Attribute `gorm:"polymorphic:Entity;polymorphic_value:tenant"`
	// END common fields

	State            int `gorm:"not null;default:0"`
	MaxApplications  *WrappedUint64
	MaxClients       *WrappedUint64
	MaxEndDevices    *WrappedUint64
	MaxGateways      *WrappedUint64
	MaxOrganizations *WrappedUint64
	MaxUsers         *WrappedUint64

	Configuration postgres.Jsonb
	Billing       postgres.Jsonb
}

func init() {
	registerModel(&Tenant{})
}

// functions to set fields from the tenant model into the tenant proto.
var tenantPBSetters = map[string]func(*ttipb.Tenant, *Tenant){
	nameField:             func(pb *ttipb.Tenant, tnt *Tenant) { pb.Name = tnt.Name },
	descriptionField:      func(pb *ttipb.Tenant, tnt *Tenant) { pb.Description = tnt.Description },
	attributesField:       func(pb *ttipb.Tenant, tnt *Tenant) { pb.Attributes = attributes(tnt.Attributes).toMap() },
	stateField:            func(pb *ttipb.Tenant, tnt *Tenant) { pb.State = ttnpb.State(tnt.State) },
	maxApplicationsField:  func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxApplications = tnt.MaxApplications.toPB() },
	maxClientsField:       func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxClients = tnt.MaxClients.toPB() },
	maxEndDevicesField:    func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxEndDevices = tnt.MaxEndDevices.toPB() },
	maxGatewaysField:      func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxGateways = tnt.MaxGateways.toPB() },
	maxOrganizationsField: func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxOrganizations = tnt.MaxOrganizations.toPB() },
	maxUsersField:         func(pb *ttipb.Tenant, tnt *Tenant) { pb.MaxUsers = tnt.MaxUsers.toPB() },
}

// functions to set fields from the tenant proto into the tenant model.
var tenantModelSetters = map[string]func(*Tenant, *ttipb.Tenant){
	nameField:        func(tnt *Tenant, pb *ttipb.Tenant) { tnt.Name = pb.Name },
	descriptionField: func(tnt *Tenant, pb *ttipb.Tenant) { tnt.Description = pb.Description },
	attributesField: func(tnt *Tenant, pb *ttipb.Tenant) {
		tnt.Attributes = attributes(tnt.Attributes).updateFromMap(pb.Attributes)
	},
	stateField:            func(tnt *Tenant, pb *ttipb.Tenant) { tnt.State = int(pb.State) },
	maxApplicationsField:  func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxApplications = wrappedUint64(pb.MaxApplications) },
	maxClientsField:       func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxClients = wrappedUint64(pb.MaxClients) },
	maxEndDevicesField:    func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxEndDevices = wrappedUint64(pb.MaxEndDevices) },
	maxGatewaysField:      func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxGateways = wrappedUint64(pb.MaxGateways) },
	maxOrganizationsField: func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxOrganizations = wrappedUint64(pb.MaxOrganizations) },
	maxUsersField:         func(tnt *Tenant, pb *ttipb.Tenant) { tnt.MaxUsers = wrappedUint64(pb.MaxUsers) },
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultTenantFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(tenantPBSetters))
	for path := range tenantPBSetters {
		paths = append(paths, path)
	}
	paths = append(paths, "configuration", "billing")
	defaultTenantFieldMask.Paths = paths
}

// fieldmask path to column name in tenants table.
var tenantColumnNames = map[string][]string{
	attributesField:       {},
	billingField:          {billingField},
	contactInfoField:      {},
	nameField:             {nameField},
	descriptionField:      {descriptionField},
	stateField:            {stateField},
	maxApplicationsField:  {maxApplicationsField},
	maxClientsField:       {maxClientsField},
	maxEndDevicesField:    {maxEndDevicesField},
	maxGatewaysField:      {maxGatewaysField},
	maxOrganizationsField: {maxOrganizationsField},
	maxUsersField:         {maxUsersField},
	configurationField:    {configurationField},
}

func (tnt Tenant) toPB(pb *ttipb.Tenant, fieldMask *types.FieldMask) error {
	pb.TenantIdentifiers.TenantID = tnt.TenantID
	pb.CreatedAt = cleanTime(tnt.CreatedAt)
	pb.UpdatedAt = cleanTime(tnt.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultTenantFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := tenantPBSetters[path]; ok {
			setter(pb, &tnt)
		}
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), configurationField) {
		if len(tnt.Configuration.RawMessage) > 0 {
			var configuration ttipb.Configuration
			if err := json.Unmarshal(tnt.Configuration.RawMessage, &configuration); err != nil {
				return err
			}
			pb.Configuration = &configuration
		} else {
			pb.Configuration = nil
		}
		if err := pb.SetFields(pb, fieldMask.Paths...); err != nil {
			return err
		}
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), billingField) {
		if len(tnt.Billing.RawMessage) > 0 {
			var billing ttipb.Billing
			if err := jsonpb.TTN().Unmarshal(tnt.Billing.RawMessage, &billing); err != nil {
				return err
			}
			pb.Billing = &billing
		} else {
			pb.Billing = nil
		}
		if err := pb.SetFields(pb, fieldMask.Paths...); err != nil {
			return err
		}
	}
	return nil
}

func (tnt *Tenant) fromPB(pb *ttipb.Tenant, fieldMask *types.FieldMask) (columns []string, err error) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultTenantFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := tenantModelSetters[path]; ok {
			setter(tnt, pb)
			if columnNames, ok := tenantColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			} else {
				columns = append(columns, path)
			}
			continue
		}
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), configurationField) {
		var configuration ttipb.Configuration
		if len(tnt.Configuration.RawMessage) > 0 {
			if err = json.Unmarshal(tnt.Configuration.RawMessage, &configuration); err != nil {
				return nil, err
			}
		}
		tmp := &ttipb.Tenant{Configuration: &configuration}
		if err := tmp.SetFields(pb, fieldMask.Paths...); err != nil {
			return nil, err
		}
		if tmp.Configuration == nil {
			tnt.Configuration.RawMessage = nil
		} else if tnt.Configuration.RawMessage, err = json.Marshal(tmp.Configuration); err != nil {
			return nil, err
		}
		columns = append(columns, configurationField)
	}
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask.Paths), billingField) {
		var billing ttipb.Billing
		if len(tnt.Billing.RawMessage) > 0 {
			if err = jsonpb.TTN().Unmarshal(tnt.Billing.RawMessage, &billing); err != nil {
				return nil, err
			}
		}
		tmp := &ttipb.Tenant{Billing: &billing}
		if err := tmp.SetFields(pb, fieldMask.Paths...); err != nil {
			return nil, err
		}
		if tmp.Billing == nil {
			tnt.Billing.RawMessage = nil
		} else if tnt.Billing.RawMessage, err = jsonpb.TTN().Marshal(tmp.Billing); err != nil {
			return nil, err
		}
		columns = append(columns, billingField)
	}
	return
}
