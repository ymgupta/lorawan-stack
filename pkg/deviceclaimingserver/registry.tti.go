// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// AuthorizedApplicationRegistry is a store for applications.
type AuthorizedApplicationRegistry interface {
	// Get returns the authorized application by its identifiers.
	Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error)
	// Set creates, updates or deletes the application by its identifiers.
	Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string, f func(*ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error)) (*ttipb.ApplicationAPIKey, error)
}
