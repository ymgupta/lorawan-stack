// Copyright © 2019 The Things Industries B.V.

package ttipb

// BuildLicenseKey builds a LicenseKey for the License.
func (m *License) BuildLicenseKey() (*LicenseKey, error) {
	var k LicenseKey
	if err := k.MarshalLicense(m); err != nil {
		return nil, err
	}
	return &k, nil
}

// Fields implements log.Fielder.
func (m *License) Fields() map[string]interface{} {
	fields := make(map[string]interface{})
	if id := m.GetLicenseID(); id != "" {
		fields["license_id"] = id
	}
	if issuer := m.GetLicenseIssuerID(); issuer != "" {
		fields["issuer"] = issuer
	}
	if !m.ValidFrom.IsZero() {
		fields["valid_from"] = m.ValidFrom
	}
	if !m.ValidUntil.IsZero() {
		fields["valid_until"] = m.ValidUntil
	}
	if m.MinVersion != "" {
		fields["min_version"] = m.MinVersion
	}
	if m.MaxVersion != "" {
		fields["max_version"] = m.MaxVersion
	}
	if len(m.Components) > 0 {
		fields["components"] = m.Components
	}
	if len(m.ComponentAddressRegexps) > 0 {
		fields["component_address_regexps"] = m.ComponentAddressRegexps
	}
	if len(m.DevAddrPrefixes) > 0 {
		fields["dev_addr_prefixes"] = m.DevAddrPrefixes
	}
	if len(m.JoinEUIPrefixes) > 0 {
		fields["join_eui_prefixes"] = m.JoinEUIPrefixes
	}
	if m.MultiTenancy {
		fields["multi_tenancy"] = true
	}
	if m.MaxApplications != nil {
		fields["max_applications"] = m.MaxApplications.Value
	}
	if m.MaxClients != nil {
		fields["max_clients"] = m.MaxClients.Value
	}
	if m.MaxEndDevices != nil {
		fields["max_end_devices"] = m.MaxEndDevices.Value
	}
	if m.MaxGateways != nil {
		fields["max_gateways"] = m.MaxGateways.Value
	}
	if m.MaxOrganizations != nil {
		fields["max_organizations"] = m.MaxOrganizations.Value
	}
	if m.MaxUsers != nil {
		fields["max_users"] = m.MaxUsers.Value
	}
	if m.Metering != nil {
		switch m.Metering.Metering.(type) {
		case *MeteringConfiguration_AWS_:
			fields["metering"] = "AWS"
		}
	}
	return fields
}

// MarshalLicense marshals the license into the license field of the LicenseKey.
func (m *LicenseKey) MarshalLicense(l *License) error {
	licenseBytes, err := l.Marshal()
	if err != nil {
		return err
	}
	m.License = licenseBytes
	return nil
}

// UnmarshalLicense unmarshals the license from the license field of the LicenseKey.
func (m *LicenseKey) UnmarshalLicense() (*License, error) {
	if m == nil {
		return nil, nil
	}
	var license License
	if err := license.Unmarshal(m.License); err != nil {
		return nil, err
	}
	return &license, nil
}
