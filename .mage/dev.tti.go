// Copyright © 2020 The Things Industries B.V.

package ttnmage

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

// InitMTStack initializes the Stack with multi-tenancy.
func (Dev) InitMTStack() error {
	tenandID := "thethings"
	if mg.Verbose() {
		fmt.Printf("Initializing the Stack with multi-tenancy\n")
	}
	if err := execGo("run", "./cmd/tti-lw-stack", "is-db", "init"); err != nil {
		return err
	}
	if err := execGo("run", "./cmd/tti-lw-stack", "is-db", "create-tenant", "--id", tenandID); err != nil {
		return err
	}
	if err := execGo("run", "./cmd/tti-lw-stack", "is-db", "create-admin-user",
		"--tenant-id", tenandID,
		"--id", "admin",
		"--email", "admin@localhost",
		"--password", "admin",
	); err != nil {
		return err
	}
	if err := execGo("run", "./cmd/tti-lw-stack", "is-db", "create-oauth-client",
		"--tenant-id", "NULL",
		"--id", "cli",
		"--name", "Command Line Interface",
		"--no-secret",
		"--redirect-uri", "local-callback",
		"--redirect-uri", "code",
	); err != nil {
		return err
	}
	if err := execGo("run", "./cmd/tti-lw-stack", "is-db", "create-oauth-client",
		"--tenant-id", "NULL",
		"--id", "console",
		"--name", "Console",
		"--secret", "console",
		"--redirect-uri", "https://localhost:8885/console/oauth/callback",
		"--redirect-uri", "http://localhost:1885/console/oauth/callback",
		"--redirect-uri", "/console/oauth/callback",
		"--logout-redirect-uri", "https://localhost:8885/console",
		"--logout-redirect-uri", "http://localhost:1885/console",
		"--logout-redirect-uri", "/console",
	); err != nil {
		return err
	}
	return execGo("run", "./cmd/tti-lw-stack", "is-db", "create-oauth-client",
		"--tenant-id", "NULL",
		"--id", "device-claiming",
		"--name", "Device Claiming",
		"--secret", "device-claiming",
		"--redirect-uri", "https://localhost:8885/claim/oauth/callback",
		"--redirect-uri", "http://localhost:1885/claim/oauth/callback",
		"--redirect-uri", "/claim/oauth/callback",
		"--logout-redirect-uri", "https://localhost:8885/claim",
		"--logout-redirect-uri", "http://localhost:1885/claim",
		"--logout-redirect-uri", "/claim",
	)
}
