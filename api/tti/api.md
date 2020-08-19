<a name="top"></a>

# API Documentation

## <a name="toc">Table of Contents</a>

- [File `lorawan-stack/api/tti/application_api_key.proto`](#lorawan-stack/api/tti/application_api_key.proto)
  - [Message `ApplicationAPIKey`](#tti.lorawan.v3.ApplicationAPIKey)
- [File `lorawan-stack/api/tti/authentication_provider.proto`](#lorawan-stack/api/tti/authentication_provider.proto)
  - [Message `AuthenticationProvider`](#tti.lorawan.v3.AuthenticationProvider)
  - [Message `AuthenticationProvider.Configuration`](#tti.lorawan.v3.AuthenticationProvider.Configuration)
  - [Message `AuthenticationProvider.OIDC`](#tti.lorawan.v3.AuthenticationProvider.OIDC)
  - [Message `AuthenticationProviderIdentifiers`](#tti.lorawan.v3.AuthenticationProviderIdentifiers)
- [File `lorawan-stack/api/tti/billing.proto`](#lorawan-stack/api/tti/billing.proto)
  - [Message `Billing`](#tti.lorawan.v3.Billing)
  - [Message `Billing.Stripe`](#tti.lorawan.v3.Billing.Stripe)
- [File `lorawan-stack/api/tti/configuration.proto`](#lorawan-stack/api/tti/configuration.proto)
  - [Message `Configuration`](#tti.lorawan.v3.Configuration)
  - [Message `Configuration.Cluster`](#tti.lorawan.v3.Configuration.Cluster)
  - [Message `Configuration.Cluster.IdentityServer`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer)
  - [Message `Configuration.Cluster.IdentityServer.EndDevicePicture`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.EndDevicePicture)
  - [Message `Configuration.Cluster.IdentityServer.ProfilePicture`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.ProfilePicture)
  - [Message `Configuration.Cluster.IdentityServer.UserRegistration`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration)
  - [Message `Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval)
  - [Message `Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation)
  - [Message `Configuration.Cluster.IdentityServer.UserRegistration.Invitation`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.Invitation)
  - [Message `Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements)
  - [Message `Configuration.Cluster.IdentityServer.UserRights`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRights)
  - [Message `Configuration.Cluster.NetworkServer`](#tti.lorawan.v3.Configuration.Cluster.NetworkServer)
  - [Message `Configuration.UI`](#tti.lorawan.v3.Configuration.UI)
- [File `lorawan-stack/api/tti/external_user.proto`](#lorawan-stack/api/tti/external_user.proto)
  - [Message `ExternalUser`](#tti.lorawan.v3.ExternalUser)
- [File `lorawan-stack/api/tti/identifiers.proto`](#lorawan-stack/api/tti/identifiers.proto)
  - [Message `LicenseIdentifiers`](#tti.lorawan.v3.LicenseIdentifiers)
  - [Message `LicenseIssuerIdentifiers`](#tti.lorawan.v3.LicenseIssuerIdentifiers)
  - [Message `TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers)
- [File `lorawan-stack/api/tti/license.proto`](#lorawan-stack/api/tti/license.proto)
  - [Message `License`](#tti.lorawan.v3.License)
  - [Message `LicenseKey`](#tti.lorawan.v3.LicenseKey)
  - [Message `LicenseKey.Signature`](#tti.lorawan.v3.LicenseKey.Signature)
  - [Message `LicenseUpdate`](#tti.lorawan.v3.LicenseUpdate)
  - [Message `MeteringConfiguration`](#tti.lorawan.v3.MeteringConfiguration)
  - [Message `MeteringConfiguration.AWS`](#tti.lorawan.v3.MeteringConfiguration.AWS)
  - [Message `MeteringConfiguration.Prometheus`](#tti.lorawan.v3.MeteringConfiguration.Prometheus)
  - [Message `MeteringData`](#tti.lorawan.v3.MeteringData)
  - [Message `MeteringData.TenantMeteringData`](#tti.lorawan.v3.MeteringData.TenantMeteringData)
- [File `lorawan-stack/api/tti/tenant.proto`](#lorawan-stack/api/tti/tenant.proto)
  - [Message `CreateTenantRequest`](#tti.lorawan.v3.CreateTenantRequest)
  - [Message `GetTenantIdentifiersForEndDeviceEUIsRequest`](#tti.lorawan.v3.GetTenantIdentifiersForEndDeviceEUIsRequest)
  - [Message `GetTenantIdentifiersForGatewayEUIRequest`](#tti.lorawan.v3.GetTenantIdentifiersForGatewayEUIRequest)
  - [Message `GetTenantRegistryTotalsRequest`](#tti.lorawan.v3.GetTenantRegistryTotalsRequest)
  - [Message `GetTenantRequest`](#tti.lorawan.v3.GetTenantRequest)
  - [Message `ListTenantsRequest`](#tti.lorawan.v3.ListTenantsRequest)
  - [Message `Tenant`](#tti.lorawan.v3.Tenant)
  - [Message `Tenant.AttributesEntry`](#tti.lorawan.v3.Tenant.AttributesEntry)
  - [Message `TenantRegistryTotals`](#tti.lorawan.v3.TenantRegistryTotals)
  - [Message `Tenants`](#tti.lorawan.v3.Tenants)
  - [Message `UpdateTenantRequest`](#tti.lorawan.v3.UpdateTenantRequest)
- [File `lorawan-stack/api/tti/tenant_services.proto`](#lorawan-stack/api/tti/tenant_services.proto)
  - [Service `TenantRegistry`](#tti.lorawan.v3.TenantRegistry)
- [File `lorawan-stack/api/tti/tenantbillingserver.proto`](#lorawan-stack/api/tti/tenantbillingserver.proto)
  - [Service `Tbs`](#tti.lorawan.v3.Tbs)
- [Scalar Value Types](#scalar-value-types)

## <a name="lorawan-stack/api/tti/application_api_key.proto">File `lorawan-stack/api/tti/application_api_key.proto`</a>

### <a name="tti.lorawan.v3.ApplicationAPIKey">Message `ApplicationAPIKey`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ttn.lorawan.v3.ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `api_key` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`string.min_len`: `1`</p> |

## <a name="lorawan-stack/api/tti/authentication_provider.proto">File `lorawan-stack/api/tti/authentication_provider.proto`</a>

### <a name="tti.lorawan.v3.AuthenticationProvider">Message `AuthenticationProvider`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`AuthenticationProviderIdentifiers`](#tti.lorawan.v3.AuthenticationProviderIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `name` | [`string`](#string) |  |  |
| `allow_registrations` | [`bool`](#bool) |  |  |
| `configuration` | [`AuthenticationProvider.Configuration`](#tti.lorawan.v3.AuthenticationProvider.Configuration) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |

### <a name="tti.lorawan.v3.AuthenticationProvider.Configuration">Message `AuthenticationProvider.Configuration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `oidc` | [`AuthenticationProvider.OIDC`](#tti.lorawan.v3.AuthenticationProvider.OIDC) |  |  |

### <a name="tti.lorawan.v3.AuthenticationProvider.OIDC">Message `AuthenticationProvider.OIDC`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [`string`](#string) |  |  |
| `client_secret` | [`string`](#string) |  |  |
| `provider_url` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `provider_url` | <p>`string.uri`: `true`</p> |

### <a name="tti.lorawan.v3.AuthenticationProviderIdentifiers">Message `AuthenticationProviderIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `provider_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `provider_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

## <a name="lorawan-stack/api/tti/billing.proto">File `lorawan-stack/api/tti/billing.proto`</a>

### <a name="tti.lorawan.v3.Billing">Message `Billing`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `stripe` | [`Billing.Stripe`](#tti.lorawan.v3.Billing.Stripe) |  |  |

### <a name="tti.lorawan.v3.Billing.Stripe">Message `Billing.Stripe`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `customer_id` | [`string`](#string) |  |  |
| `plan_id` | [`string`](#string) |  |  |
| `subscription_id` | [`string`](#string) |  |  |
| `subscription_item_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `customer_id` | <p>`string.min_len`: `1`</p> |
| `plan_id` | <p>`string.min_len`: `1`</p> |
| `subscription_id` | <p>`string.min_len`: `1`</p> |
| `subscription_item_id` | <p>`string.min_len`: `1`</p> |

## <a name="lorawan-stack/api/tti/configuration.proto">File `lorawan-stack/api/tti/configuration.proto`</a>

### <a name="tti.lorawan.v3.Configuration">Message `Configuration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_cluster` | [`Configuration.Cluster`](#tti.lorawan.v3.Configuration.Cluster) |  | Default cluster configuration. |

### <a name="tti.lorawan.v3.Configuration.Cluster">Message `Configuration.Cluster`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ui` | [`Configuration.UI`](#tti.lorawan.v3.Configuration.UI) |  |  |
| `is` | [`Configuration.Cluster.IdentityServer`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer) |  | Identity Server configuration. |
| `ns` | [`Configuration.Cluster.NetworkServer`](#tti.lorawan.v3.Configuration.Cluster.NetworkServer) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer">Message `Configuration.Cluster.IdentityServer`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_registration` | [`Configuration.Cluster.IdentityServer.UserRegistration`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration) |  |  |
| `profile_picture` | [`Configuration.Cluster.IdentityServer.ProfilePicture`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.ProfilePicture) |  |  |
| `end_device_picture` | [`Configuration.Cluster.IdentityServer.EndDevicePicture`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.EndDevicePicture) |  |  |
| `user_rights` | [`Configuration.Cluster.IdentityServer.UserRights`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRights) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.EndDevicePicture">Message `Configuration.Cluster.IdentityServer.EndDevicePicture`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `disable_upload` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.ProfilePicture">Message `Configuration.Cluster.IdentityServer.ProfilePicture`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `disable_upload` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `use_gravatar` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration">Message `Configuration.Cluster.IdentityServer.UserRegistration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `invitation` | [`Configuration.Cluster.IdentityServer.UserRegistration.Invitation`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.Invitation) |  |  |
| `contact_info_validation` | [`Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation) |  |  |
| `admin_approval` | [`Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval) |  |  |
| `password_requirements` | [`Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements`](#tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval">Message `Configuration.Cluster.IdentityServer.UserRegistration.AdminApproval`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation">Message `Configuration.Cluster.IdentityServer.UserRegistration.ContactInfoValidation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.Invitation">Message `Configuration.Cluster.IdentityServer.UserRegistration.Invitation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `token_ttl` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements">Message `Configuration.Cluster.IdentityServer.UserRegistration.PasswordRequirements`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_length` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `max_length` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_uppercase` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_digits` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_special` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.IdentityServer.UserRights">Message `Configuration.Cluster.IdentityServer.UserRights`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `create_applications` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_clients` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_gateways` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_organizations` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="tti.lorawan.v3.Configuration.Cluster.NetworkServer">Message `Configuration.Cluster.NetworkServer`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr_prefixes` | [`bytes`](#bytes) | repeated |  |
| `deduplication_window` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `cooldown_window` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

### <a name="tti.lorawan.v3.Configuration.UI">Message `Configuration.UI`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `branding_base_url` | [`string`](#string) |  |  |

## <a name="lorawan-stack/api/tti/external_user.proto">File `lorawan-stack/api/tti/external_user.proto`</a>

### <a name="tti.lorawan.v3.ExternalUser">Message `ExternalUser`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`ttn.lorawan.v3.UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `provider_ids` | [`AuthenticationProviderIdentifiers`](#tti.lorawan.v3.AuthenticationProviderIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `external_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `provider_ids` | <p>`message.required`: `true`</p> |

## <a name="lorawan-stack/api/tti/identifiers.proto">File `lorawan-stack/api/tti/identifiers.proto`</a>

### <a name="tti.lorawan.v3.LicenseIdentifiers">Message `LicenseIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `license_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `license_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="tti.lorawan.v3.LicenseIssuerIdentifiers">Message `LicenseIssuerIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `license_issuer_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `license_issuer_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="tti.lorawan.v3.TenantIdentifiers">Message `TenantIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

## <a name="lorawan-stack/api/tti/license.proto">File `lorawan-stack/api/tti/license.proto`</a>

### <a name="tti.lorawan.v3.License">Message `License`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`LicenseIdentifiers`](#tti.lorawan.v3.LicenseIdentifiers) |  | Immutable and unique public identifier for the License. Generated by the License Server. |
| `license_issuer_ids` | [`LicenseIssuerIdentifiers`](#tti.lorawan.v3.LicenseIssuerIdentifiers) |  | Issuer of the license. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `valid_from` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | The license is not valid before this time. |
| `valid_until` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | The license is not valid after this time. |
| `warn_for` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | For how long (before valid_until) to warn about license expiry. |
| `limit_for` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | For how long (before valid_until) to limit non-critical functionality. |
| `min_version` | [`string`](#string) |  | If set, sets the minimum version allowed by this license (major.minor.patch). |
| `max_version` | [`string`](#string) |  | If set, sets the maximum version allowed by this license (major.minor.patch). |
| `components` | [`ttn.lorawan.v3.ClusterRole`](#ttn.lorawan.v3.ClusterRole) | repeated | If set, only the given components can be started. |
| `component_address_regexps` | [`string`](#string) | repeated | If set, the server addresses must match any of these regexps. |
| `dev_addr_prefixes` | [`bytes`](#bytes) | repeated | If set, the configured DevAddr prefixes must match any of these prefixes. |
| `join_eui_prefixes` | [`bytes`](#bytes) | repeated | If set, the configured JoinEUI prefixes must match any of these prefixes. |
| `multi_tenancy` | [`bool`](#bool) |  | Indicates whether multi-tenancy support is included. |
| `max_applications` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of applications that can be created. |
| `max_clients` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of clients that can be created. |
| `max_end_devices` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of end_devices that can be created. |
| `max_gateways` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of gateways that can be created. |
| `max_organizations` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of organizations that can be created. |
| `max_users` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of users that can be created. |
| `metering` | [`MeteringConfiguration`](#tti.lorawan.v3.MeteringConfiguration) |  | If set, requires checking in with a metering service. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `id` | <p>`message.required`: `true`</p> |
| `license_issuer_ids` | <p>`message.required`: `true`</p> |

### <a name="tti.lorawan.v3.LicenseKey">Message `LicenseKey`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `license` | [`bytes`](#bytes) |  | The marshaled License message. |
| `signatures` | [`LicenseKey.Signature`](#tti.lorawan.v3.LicenseKey.Signature) | repeated | Signatures for the license bytes. The LicenseKey is invalid if it does not contain any signature with a known key_id or if it contains any invalid signature. |

### <a name="tti.lorawan.v3.LicenseKey.Signature">Message `LicenseKey.Signature`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [`string`](#string) |  | The ID of the key used to sign license. |
| `signature` | [`bytes`](#bytes) |  | Signature for license using the key identified by key_id. |

### <a name="tti.lorawan.v3.LicenseUpdate">Message `LicenseUpdate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `extend_valid_until` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | How long the license validity should be extended (relative to the current time) on update. |

### <a name="tti.lorawan.v3.MeteringConfiguration">Message `MeteringConfiguration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `interval` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | How frequently to report to the metering service. |
| `on_success` | [`LicenseUpdate`](#tti.lorawan.v3.LicenseUpdate) |  | How to update the license on success. |
| `aws` | [`MeteringConfiguration.AWS`](#tti.lorawan.v3.MeteringConfiguration.AWS) |  |  |
| `prometheus` | [`MeteringConfiguration.Prometheus`](#tti.lorawan.v3.MeteringConfiguration.Prometheus) |  |  |

### <a name="tti.lorawan.v3.MeteringConfiguration.AWS">Message `MeteringConfiguration.AWS`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sku` | [`string`](#string) |  |  |

### <a name="tti.lorawan.v3.MeteringConfiguration.Prometheus">Message `MeteringConfiguration.Prometheus`</a>

### <a name="tti.lorawan.v3.MeteringData">Message `MeteringData`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenants` | [`MeteringData.TenantMeteringData`](#tti.lorawan.v3.MeteringData.TenantMeteringData) | repeated |  |

### <a name="tti.lorawan.v3.MeteringData.TenantMeteringData">Message `MeteringData.TenantMeteringData`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant_id` | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |  |
| `totals` | [`TenantRegistryTotals`](#tti.lorawan.v3.TenantRegistryTotals) |  |  |

## <a name="lorawan-stack/api/tti/tenant.proto">File `lorawan-stack/api/tti/tenant.proto`</a>

### <a name="tti.lorawan.v3.CreateTenantRequest">Message `CreateTenantRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant` | [`Tenant`](#tti.lorawan.v3.Tenant) |  |  |
| `initial_user` | [`ttn.lorawan.v3.User`](#ttn.lorawan.v3.User) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant` | <p>`message.required`: `true`</p> |

### <a name="tti.lorawan.v3.GetTenantIdentifiersForEndDeviceEUIsRequest">Message `GetTenantIdentifiersForEndDeviceEUIsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  |  |

### <a name="tti.lorawan.v3.GetTenantIdentifiersForGatewayEUIRequest">Message `GetTenantIdentifiersForGatewayEUIRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `eui` | [`bytes`](#bytes) |  |  |

### <a name="tti.lorawan.v3.GetTenantRegistryTotalsRequest">Message `GetTenantRegistryTotalsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant_ids` | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

### <a name="tti.lorawan.v3.GetTenantRequest">Message `GetTenantRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant_ids` | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant_ids` | <p>`message.required`: `true`</p> |

### <a name="tti.lorawan.v3.ListTenantsRequest">Message `ListTenantsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ tenant_id -tenant_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="tti.lorawan.v3.Tenant">Message `Tenant`</a>

Tenant is the message that defines a Tenant in the network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `name` | [`string`](#string) |  |  |
| `description` | [`string`](#string) |  |  |
| `attributes` | [`Tenant.AttributesEntry`](#tti.lorawan.v3.Tenant.AttributesEntry) | repeated |  |
| `contact_info` | [`ttn.lorawan.v3.ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| `max_applications` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of applications that can be created. |
| `max_clients` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of clients that can be created. |
| `max_end_devices` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of end_devices that can be created. |
| `max_gateways` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of gateways that can be created. |
| `max_organizations` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of organizations that can be created. |
| `max_users` | [`google.protobuf.UInt64Value`](#google.protobuf.UInt64Value) |  | If set, restricts the maximum number of users that can be created. |
| `state` | [`ttn.lorawan.v3.State`](#ttn.lorawan.v3.State) |  | The reviewing state of the tenant. This field can only be modified by tenant admins. |
| `capabilities` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |
| `configuration` | [`Configuration`](#tti.lorawan.v3.Configuration) |  |  |
| `billing` | [`Billing`](#tti.lorawan.v3.Billing) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `state` | <p>`enum.defined_only`: `true`</p> |

### <a name="tti.lorawan.v3.Tenant.AttributesEntry">Message `Tenant.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="tti.lorawan.v3.TenantRegistryTotals">Message `TenantRegistryTotals`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `applications` | [`uint64`](#uint64) |  |  |
| `clients` | [`uint64`](#uint64) |  |  |
| `end_devices` | [`uint64`](#uint64) |  |  |
| `gateways` | [`uint64`](#uint64) |  |  |
| `organizations` | [`uint64`](#uint64) |  |  |
| `users` | [`uint64`](#uint64) |  |  |

### <a name="tti.lorawan.v3.Tenants">Message `Tenants`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenants` | [`Tenant`](#tti.lorawan.v3.Tenant) | repeated |  |

### <a name="tti.lorawan.v3.UpdateTenantRequest">Message `UpdateTenantRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant` | [`Tenant`](#tti.lorawan.v3.Tenant) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant` | <p>`message.required`: `true`</p> |

## <a name="lorawan-stack/api/tti/tenant_services.proto">File `lorawan-stack/api/tti/tenant_services.proto`</a>

### <a name="tti.lorawan.v3.TenantRegistry">Service `TenantRegistry`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateTenantRequest`](#tti.lorawan.v3.CreateTenantRequest) | [`Tenant`](#tti.lorawan.v3.Tenant) |  |
| `Get` | [`GetTenantRequest`](#tti.lorawan.v3.GetTenantRequest) | [`Tenant`](#tti.lorawan.v3.Tenant) |  |
| `GetRegistryTotals` | [`GetTenantRegistryTotalsRequest`](#tti.lorawan.v3.GetTenantRegistryTotalsRequest) | [`TenantRegistryTotals`](#tti.lorawan.v3.TenantRegistryTotals) |  |
| `List` | [`ListTenantsRequest`](#tti.lorawan.v3.ListTenantsRequest) | [`Tenants`](#tti.lorawan.v3.Tenants) |  |
| `Update` | [`UpdateTenantRequest`](#tti.lorawan.v3.UpdateTenantRequest) | [`Tenant`](#tti.lorawan.v3.Tenant) |  |
| `Delete` | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) |  |
| `GetIdentifiersForEndDeviceEUIs` | [`GetTenantIdentifiersForEndDeviceEUIsRequest`](#tti.lorawan.v3.GetTenantIdentifiersForEndDeviceEUIsRequest) | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |
| `GetIdentifiersForGatewayEUI` | [`GetTenantIdentifiersForGatewayEUIRequest`](#tti.lorawan.v3.GetTenantIdentifiersForGatewayEUIRequest) | [`TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/tenants` | `*` |
| `Get` | `GET` | `/api/v3/tenants/{tenant_ids.tenant_id}` |  |
| `GetRegistryTotals` | `GET` | `/api/v3/tenants/{tenant_ids.tenant_id}/registry-totals` |  |
| `List` | `GET` | `/api/v3/tenants` |  |
| `Update` | `PUT` | `/api/v3/tenants/{tenant.ids.tenant_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/tenants/{tenant_id}` |  |

## <a name="lorawan-stack/api/tti/tenantbillingserver.proto">File `lorawan-stack/api/tti/tenantbillingserver.proto`</a>

### <a name="tti.lorawan.v3.Tbs">Service `Tbs`</a>

The Tbs service manages the Tenant Billing Server metering reporting.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Report` | [`MeteringData`](#tti.lorawan.v3.MeteringData) | [`.google.protobuf.Empty`](#google.protobuf.Empty) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Report` | `POST` | `/api/v3/tbs/report` |  |

## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double" /> double |  | double | double | float |
| <a name="float" /> float |  | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |
