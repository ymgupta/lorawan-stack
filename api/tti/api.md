<a name="top"></a>

# API Documentation

## <a name="toc">Table of Contents</a>

- [File `lorawan-stack/api/tti/identifiers.proto`](#lorawan-stack/api/tti/identifiers.proto)
  - [Message `TenantIdentifiers`](#tti.lorawan.v3.TenantIdentifiers)
- [File `lorawan-stack/api/tti/tenant.proto`](#lorawan-stack/api/tti/tenant.proto)
  - [Message `CreateTenantRequest`](#tti.lorawan.v3.CreateTenantRequest)
  - [Message `GetTenantRequest`](#tti.lorawan.v3.GetTenantRequest)
  - [Message `ListTenantsRequest`](#tti.lorawan.v3.ListTenantsRequest)
  - [Message `Tenant`](#tti.lorawan.v3.Tenant)
  - [Message `Tenant.AttributesEntry`](#tti.lorawan.v3.Tenant.AttributesEntry)
  - [Message `Tenants`](#tti.lorawan.v3.Tenants)
  - [Message `UpdateTenantRequest`](#tti.lorawan.v3.UpdateTenantRequest)
- [File `lorawan-stack/api/tti/tenant_services.proto`](#lorawan-stack/api/tti/tenant_services.proto)
  - [Message `GetTenantIdentifiersForEndDeviceEUIsRequest`](#tti.lorawan.v3.GetTenantIdentifiersForEndDeviceEUIsRequest)
  - [Message `GetTenantIdentifiersForGatewayEUIRequest`](#tti.lorawan.v3.GetTenantIdentifiersForGatewayEUIRequest)
  - [Message `GetTenantRegistryTotalsRequest`](#tti.lorawan.v3.GetTenantRegistryTotalsRequest)
  - [Message `TenantRegistryTotals`](#tti.lorawan.v3.TenantRegistryTotals)
  - [Service `TenantRegistry`](#tti.lorawan.v3.TenantRegistry)
- [Scalar Value Types](#scalar-value-types)

## <a name="lorawan-stack/api/tti/identifiers.proto">File `lorawan-stack/api/tti/identifiers.proto`</a>

### <a name="tti.lorawan.v3.TenantIdentifiers">Message `TenantIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

## <a name="lorawan-stack/api/tti/tenant.proto">File `lorawan-stack/api/tti/tenant.proto`</a>

### <a name="tti.lorawan.v3.CreateTenantRequest">Message `CreateTenantRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tenant` | [`Tenant`](#tti.lorawan.v3.Tenant) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant` | <p>`message.required`: `true`</p> |

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
| `state` | [`ttn.lorawan.v3.State`](#ttn.lorawan.v3.State) |  | The reviewing state of the tenant. This field can only be modified by tenant admins. |
| `capabilities` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |

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

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant_ids` | <p>`message.required`: `true`</p> |

### <a name="tti.lorawan.v3.TenantRegistryTotals">Message `TenantRegistryTotals`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `applications` | [`uint64`](#uint64) |  |  |
| `clients` | [`uint64`](#uint64) |  |  |
| `end_devices` | [`uint64`](#uint64) |  |  |
| `gateways` | [`uint64`](#uint64) |  |  |
| `organizations` | [`uint64`](#uint64) |  |  |
| `users` | [`uint64`](#uint64) |  |  |

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
