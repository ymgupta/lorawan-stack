// Copyright © 2019 The Things Industries B.V.

package ttnmage

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
)

// BuildTTI runs all necessary commands to build the console bundles and files in TTI configuration.
func (js Js) BuildTTI() {
	mg.SerialDeps(js.Deps, JsSDK.Build, js.BuildDll, js.BuildMainTTI)
}

// BuildMainTTI runs the webpack command with the project config in TTI configuration.
func (js Js) BuildMainTTI() error {
	mg.Deps(js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack")
	}
	webpack, err := js.webpack()
	if err != nil {
		return err
	}
	return webpack("--config", "config/webpack.config.tti.babel.js")
}

// ServeTTI builds necessary bundles and serves the console for development in TTI configuration.
func (js Js) ServeTTI() {
	mg.Deps(js.ServeMainTTI)
}

// ServeMainTTI runs webpack-dev-server in TTI configuration.
func (js Js) ServeMainTTI() error {
	mg.Deps(js.Translations, js.BackendTranslations, js.BuildDll)
	if mg.Verbose() {
		fmt.Println("Running Webpack for Main Bundle in watch mode...")
	}
	webpackServe, err := js.webpackServe()
	if err != nil {
		return err
	}
	os.Setenv("DEV_SERVER_BUILD", "true")
	return webpackServe("--config", "config/webpack.config.tti.babel.js", "-w")
}
