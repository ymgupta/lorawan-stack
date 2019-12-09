// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttipb

var TenantFieldPathsNested = []string{
	"attributes",
	"capabilities",
	"configuration",
	"configuration.default_cluster",
	"configuration.default_cluster.ui",
	"configuration.default_cluster.ui.branding_base_url",
	"contact_info",
	"created_at",
	"description",
	"ids",
	"ids.tenant_id",
	"max_applications",
	"max_clients",
	"max_end_devices",
	"max_gateways",
	"max_organizations",
	"max_users",
	"name",
	"state",
	"updated_at",
}

var TenantFieldPathsTopLevel = []string{
	"attributes",
	"capabilities",
	"configuration",
	"contact_info",
	"created_at",
	"description",
	"ids",
	"max_applications",
	"max_clients",
	"max_end_devices",
	"max_gateways",
	"max_organizations",
	"max_users",
	"name",
	"state",
	"updated_at",
}
var TenantsFieldPathsNested = []string{
	"tenants",
}

var TenantsFieldPathsTopLevel = []string{
	"tenants",
}
var GetTenantRequestFieldPathsNested = []string{
	"field_mask",
	"tenant_ids",
	"tenant_ids.tenant_id",
}

var GetTenantRequestFieldPathsTopLevel = []string{
	"field_mask",
	"tenant_ids",
}
var ListTenantsRequestFieldPathsNested = []string{
	"field_mask",
	"limit",
	"order",
	"page",
}

var ListTenantsRequestFieldPathsTopLevel = []string{
	"field_mask",
	"limit",
	"order",
	"page",
}
var CreateTenantRequestFieldPathsNested = []string{
	"initial_user",
	"tenant",
	"tenant.attributes",
	"tenant.capabilities",
	"tenant.configuration",
	"tenant.configuration.default_cluster",
	"tenant.configuration.default_cluster.ui",
	"tenant.configuration.default_cluster.ui.branding_base_url",
	"tenant.contact_info",
	"tenant.created_at",
	"tenant.description",
	"tenant.ids",
	"tenant.ids.tenant_id",
	"tenant.max_applications",
	"tenant.max_clients",
	"tenant.max_end_devices",
	"tenant.max_gateways",
	"tenant.max_organizations",
	"tenant.max_users",
	"tenant.name",
	"tenant.state",
	"tenant.updated_at",
}

var CreateTenantRequestFieldPathsTopLevel = []string{
	"initial_user",
	"tenant",
}
var UpdateTenantRequestFieldPathsNested = []string{
	"field_mask",
	"tenant",
	"tenant.attributes",
	"tenant.capabilities",
	"tenant.configuration",
	"tenant.configuration.default_cluster",
	"tenant.configuration.default_cluster.ui",
	"tenant.configuration.default_cluster.ui.branding_base_url",
	"tenant.contact_info",
	"tenant.created_at",
	"tenant.description",
	"tenant.ids",
	"tenant.ids.tenant_id",
	"tenant.max_applications",
	"tenant.max_clients",
	"tenant.max_end_devices",
	"tenant.max_gateways",
	"tenant.max_organizations",
	"tenant.max_users",
	"tenant.name",
	"tenant.state",
	"tenant.updated_at",
}

var UpdateTenantRequestFieldPathsTopLevel = []string{
	"field_mask",
	"tenant",
}
var GetTenantIdentifiersForEndDeviceEUIsRequestFieldPathsNested = []string{
	"dev_eui",
	"join_eui",
}

var GetTenantIdentifiersForEndDeviceEUIsRequestFieldPathsTopLevel = []string{
	"dev_eui",
	"join_eui",
}
var GetTenantIdentifiersForGatewayEUIRequestFieldPathsNested = []string{
	"eui",
}

var GetTenantIdentifiersForGatewayEUIRequestFieldPathsTopLevel = []string{
	"eui",
}
var GetTenantRegistryTotalsRequestFieldPathsNested = []string{
	"field_mask",
	"tenant_ids",
	"tenant_ids.tenant_id",
}

var GetTenantRegistryTotalsRequestFieldPathsTopLevel = []string{
	"field_mask",
	"tenant_ids",
}
var TenantRegistryTotalsFieldPathsNested = []string{
	"applications",
	"clients",
	"end_devices",
	"gateways",
	"organizations",
	"users",
}

var TenantRegistryTotalsFieldPathsTopLevel = []string{
	"applications",
	"clients",
	"end_devices",
	"gateways",
	"organizations",
	"users",
}
