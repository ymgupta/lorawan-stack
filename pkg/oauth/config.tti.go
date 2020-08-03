// Copyright © 2019 The Things Industries B.V.

package oauth

import "go.thethings.network/lorawan-stack/v3/pkg/oauth/oidc"

// ProviderConfig is the configuration of the federated authentication providers.
type ProvidersConfig struct {
	OIDC oidc.Config `name:"oidc" description:"OpenID Connect provider configuration"`
}
