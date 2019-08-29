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
