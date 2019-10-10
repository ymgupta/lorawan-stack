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
	Enabled bool `name:"enabled"`

	LoadKeysFunc       KeyLoaderFunc `name:"-"`
	ReloadKeysInterval time.Duration `name:"reload-keys-interval" description:"Interval to reload keys"`

	ExposeRootKeys bool `name:"expose-root-keys" description:"Expose root keys"`

	Source string `name:"source" description:"Source keys (aws, gcp)"`
	AWS    struct {
		Region string `name:"region" description:"Region"`
	} `name:"aws" description:"Amazon Web Services settings"`
	GCP struct {
		ProjectID string `name:"project-id" description:"Project ID"`
		PartKind  string `name:"part-kind" description:"Datastore part kind"`
	} `name:"gcp" description:"Google Cloud Platform settings"`
}

type impl struct {
	parentKeysMu sync.RWMutex
	parentKeys   map[string]Key
}

var errKeyLoader = errors.DefineFailedPrecondition("key_loader", "invalid key loader")

// New returns a new Microchip provisioner.
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

	ctx = log.NewContextWithField(ctx, "namespace", "cryptoserver/provisioners/microchip")
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
