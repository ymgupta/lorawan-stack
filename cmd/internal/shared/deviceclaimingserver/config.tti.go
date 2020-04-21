// Copyright © 2019 The Things Industries B.V.

package shared

import (
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/pkg/web/oauthclient"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// DefaultDeviceClaimingServerConfig is the default configuration for the Device Claiming Server.
var DefaultDeviceClaimingServerConfig = deviceclaimingserver.Config{
	OAuth: oauthclient.Config{
		AuthorizeURL: shared.DefaultOAuthPublicURL + "/authorize",
		TokenURL:     shared.DefaultOAuthPublicURL + "/token",
		ClientID:     "device-claiming",
		ClientSecret: "device-claiming",
	},
	UI: deviceclaimingserver.UIConfig{
		TemplateData: webui.TemplateData{
			SiteName:      "The Things Enterprise Stack",
			Title:         "Device Claiming Server",
			Language:      "en",
			CanonicalURL:  shared.DefaultDeviceClaimingPublicURL,
			AssetsBaseURL: shared.DefaultAssetsBaseURL,
			IconPrefix:    "claim-",
			CSSFiles:      []string{"claim.css"},
			JSFiles:       []string{"claim.js"},
		},
		FrontendConfig: deviceclaimingserver.FrontendConfig{
			StackConfig: deviceclaimingserver.StackConfig{
				IS:  webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
				DCS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
				AS:  webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
				NS:  webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
			},
		},
	},
}
