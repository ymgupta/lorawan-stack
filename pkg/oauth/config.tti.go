// Copyright © 2019 The Things Industries B.V.

package oauth

// ProviderConfig is the shared configuration between the authentication providers.
type ProviderConfig struct {
	Name               string `name:"name" description:"Public provider name"`
	AllowRegistrations bool   `name:"allow-registrations" description:"Allow clients to be registered automatically on login"`
}

func (conf ProviderConfig) IsZero() bool {
	return conf.Name == "" &&
		!conf.AllowRegistrations
}

// OIDCConfig is the configuration of the OpenID Connect authentication provider.
type OIDCConfig struct {
	ProviderConfig `name:",squash"`
	ClientID       string `name:"client-id" description:"Client ID of the OIDC client"`
	ClientSecret   string `name:"client-secret" description:"Client secret of the OIDC client"`
	RedirectURL    string `name:"redirect-url" description:"Path on the server where the OIDC client will be redirected"`
	ProviderURL    string `name:"provider-url" description:"Path of the OIDC server"`
}

func (conf OIDCConfig) IsZero() bool {
	return conf.ProviderConfig.IsZero() &&
		conf.ClientID == "" &&
		conf.ClientSecret == "" &&
		conf.RedirectURL == "" &&
		conf.ProviderURL == ""
}

// ProviderConfig is the configuration of the federated authentication providers.
type ProvidersConfig struct {
	OIDC OIDCConfig `name:"oidc" description:"OpenID Connect provider configuration"`
}
