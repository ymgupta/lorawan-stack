// Copyright © 2020 The Things Industries B.V.

package oidc

// Config is the configuration of the OpenID Connect authentication provider.
type Config struct {
	Name               string `name:"name" description:"Public provider name"`
	AllowRegistrations bool   `name:"allow-registrations" description:"Allow clients to be registered automatically on login"`
	ClientID           string `name:"client-id" description:"Client ID of the OIDC client"`
	ClientSecret       string `name:"client-secret" description:"Client secret of the OIDC client"`
	ProviderURL        string `name:"provider-url" description:"Path of the OIDC server"`
}

// IsZero checks if the config is empty.
func (conf Config) IsZero() bool {
	return conf.Name == "" &&
		!conf.AllowRegistrations &&
		conf.ClientID == "" &&
		conf.ClientSecret == "" &&
		conf.ProviderURL == ""
}
