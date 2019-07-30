// Copyright © 2019 The Things Industries B.V.

package store

import (
	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
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
}

func init() {
	registerModel(&Tenant{})
}

// functions to set fields from the tenant model into the tenant proto.
var tenantPBSetters = map[string]func(*ttipb.Tenant, *Tenant){
	nameField:        func(pb *ttipb.Tenant, tnt *Tenant) { pb.Name = tnt.Name },
	descriptionField: func(pb *ttipb.Tenant, tnt *Tenant) { pb.Description = tnt.Description },
	attributesField:  func(pb *ttipb.Tenant, tnt *Tenant) { pb.Attributes = attributes(tnt.Attributes).toMap() },
}

// functions to set fields from the tenant proto into the tenant model.
var tenantModelSetters = map[string]func(*Tenant, *ttipb.Tenant){
	nameField:        func(tnt *Tenant, pb *ttipb.Tenant) { tnt.Name = pb.Name },
	descriptionField: func(tnt *Tenant, pb *ttipb.Tenant) { tnt.Description = pb.Description },
	attributesField: func(tnt *Tenant, pb *ttipb.Tenant) {
		tnt.Attributes = attributes(tnt.Attributes).updateFromMap(pb.Attributes)
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultTenantFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(tenantPBSetters))
	for path := range tenantPBSetters {
		paths = append(paths, path)
	}
	defaultTenantFieldMask.Paths = paths
}

// fieldmask path to column name in tenants table.
var tenantColumnNames = map[string][]string{
	attributesField:  {},
	contactInfoField: {},
	nameField:        {nameField},
	descriptionField: {descriptionField},
}

func (tnt Tenant) toPB(pb *ttipb.Tenant, fieldMask *types.FieldMask) {
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
}

func (tnt *Tenant) fromPB(pb *ttipb.Tenant, fieldMask *types.FieldMask) (columns []string) {
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
	return
}
