// Copyright © 2019 The Things Industries B.V.

package ttnmage

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
)

// Build runs the webpack command with the project config in TTI configuration.
func (js Js) BuildTTI() error {
	mg.Deps(js.Deps, js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	return js.runWebpack("config/webpack.config.tti.babel.js")
}

// Serve runs webpack-dev-server in TTI configuration.
func (js Js) ServeTTI() error {
	mg.Deps(js.Deps, js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack for Main Bundle in watch mode")
	}
	os.Setenv("DEV_SERVER_BUILD", "true")
	return js.runYarnCommandV("webpack-dev-server",
		"--config", "config/webpack.config.tti.babel.js",
		"-w",
	)
}
