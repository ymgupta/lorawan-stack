// Copyright © 2019 The Things Industries B.V.

package identityserver

func init() {
	DefaultIdentityServerConfig.OAuth.UI.SiteName = "The Things Industries LoRaWAN Stack"
	DefaultIdentityServerConfig.Email.Network.Name = DefaultIdentityServerConfig.OAuth.UI.SiteName
}
