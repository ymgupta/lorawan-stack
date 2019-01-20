// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	shared_cryptoserver "go.thethings.network/lorawan-stack/cmd/internal/shared/cryptoserver"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/cryptoserver"
)

// Config for the ttn-lw-crypto-server binary.
type Config struct {
	conf.ServiceBase `name:",squash"`
	CS               cryptoserver.Config `name:"cs"`
}

// DefaultConfig contains the default config for the ttn-lw-crypto-server binary.
var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	CS:          shared_cryptoserver.DefaultCryptoServerConfig,
}
