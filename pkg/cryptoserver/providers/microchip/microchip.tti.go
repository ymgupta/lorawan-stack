// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// KeyLoaderFunc represents a function to load keys.
type KeyLoaderFunc func(ctx context.Context) (map[string]Key, error)

// Config represents Microchip crypto service configuration.
type Config struct {
	LoadKeysFunc       KeyLoaderFunc `name:"-"`
	ReloadKeysInterval time.Duration `name:"reload-keys-interval" description:"Interval to reload keys"`

	GCPProjectID string `name:"gcp-project-id" description:"Google Cloud Platform project ID"`
	GCPPartKind  string `name:"gcp-part-kind" description:"Google Cloud Platform Datastore part kind"`
}

type impl struct {
	parentKeysMu sync.RWMutex
	parentKeys   map[string]Key
}

var errKeyLoader = errors.DefineFailedPrecondition("key_loader", "invalid key loader")

// New returns a new Microchip crypto service provider.
func New(ctx context.Context, conf *Config) (cryptoservices.NetworkApplication, error) {
	loadKeysFunc := conf.LoadKeysFunc
	if loadKeysFunc == nil {
		switch {
		case conf.GCPProjectID != "":
			partKind := conf.GCPPartKind
			if partKind == "" {
				partKind = "part"
			}
			loadKeysFunc = NewGCPKeyLoader(conf.GCPProjectID, partKind)
		default:
			return nil, errKeyLoader
		}
	}
	svc := &impl{
		parentKeys: make(map[string]Key),
	}

	ctx = log.NewContextWithField(ctx, "namespace", "cryptoserver/providers/microchip")
	logger := log.FromContext(ctx)
	loadKeys := func() error {
		logger.Debug("Load keys")
		keys, err := loadKeysFunc(ctx)
		if err != nil {
			return err
		}
		logger.WithField("count", len(keys)).Debug("Loaded keys")
		svc.parentKeysMu.Lock()
		svc.parentKeys = keys
		svc.parentKeysMu.Unlock()
		return nil
	}

	if err := loadKeys(); err != nil {
		logger.WithError(err).Error("Failed to load keys")
	}

	go func() {
		interval := conf.ReloadKeysInterval
		if interval == 0 {
			interval = 5 * time.Minute
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := loadKeys(); err != nil {
					logger.WithError(err).Warn("Failed to load keys")
				}
			}
		}
	}()

	return svc, nil
}
