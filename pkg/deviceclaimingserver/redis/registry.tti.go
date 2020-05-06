// Copyright © 2019 The Things Industries B.V.

package redis

import (
	"context"
	"runtime/trace"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	errInvalidFieldmask   = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
)

func appendImplicitApplicationAPIGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 1+len(paths)),
		"application_ids",
	), paths...)
}

func applyApplicationAPIKeyFieldMask(dst, src *ttipb.ApplicationAPIKey, paths ...string) (*ttipb.ApplicationAPIKey, error) {
	if dst == nil {
		dst = &ttipb.ApplicationAPIKey{}
	}
	return dst, dst.SetFields(src, paths...)
}

// AuthorizedApplicationRegistry is a store for authorized applications.
type AuthorizedApplicationRegistry struct {
	Redis *ttnredis.Client
}

func (r *AuthorizedApplicationRegistry) appKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Get returns the authorized application by its identifiers.
func (r *AuthorizedApplicationRegistry) Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
	defer trace.StartRegion(ctx, "get application API key").End()

	pb := &ttipb.ApplicationAPIKey{}
	if err := ttnredis.GetProto(r.Redis, r.appKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyApplicationAPIKeyFieldMask(nil, pb, appendImplicitApplicationAPIGetPaths(paths...)...)
}

// Set creates, updates or deletes the application by its identifiers.
func (r *AuthorizedApplicationRegistry) Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, gets []string, f func(*ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error)) (*ttipb.ApplicationAPIKey, error) {
	defer trace.StartRegion(ctx, "set application API key").End()

	uid := unique.ID(ctx, ids)
	uk := r.appKey(uid)

	var pb *ttipb.ApplicationAPIKey
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, uk)
		stored := &ttipb.ApplicationAPIKey{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		gets = appendImplicitApplicationAPIGetPaths(gets...)

		var err error
		if stored != nil {
			pb = &ttipb.ApplicationAPIKey{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyApplicationAPIKeyFieldMask(nil, pb, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
		if err != nil {
			return err
		}
		if stored == nil && pb == nil {
			return nil
		}
		if pb != nil && len(sets) == 0 {
			pb, err = applyApplicationAPIKeyFieldMask(nil, stored, gets...)
			return err
		}

		if pb == nil && len(sets) == 0 {
			return tx.Del(uk).Err()
		}

		if pb == nil {
			pb = &ttipb.ApplicationAPIKey{}
		}

		if pb.ApplicationIDs != ids {
			return errInvalidIdentifiers.New()
		}

		updated := &ttipb.ApplicationAPIKey{}
		if stored == nil {
			if err := ttnpb.RequireFields(sets, "application_ids"); err != nil {
				return errInvalidFieldmask.WithCause(err)
			}
			updated, err = applyApplicationAPIKeyFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}
			if updated.ApplicationIDs != ids {
				return errInvalidIdentifiers.New()
			}
		} else {
			if err := ttnpb.ProhibitFields(sets, "application_ids"); err != nil {
				return errInvalidFieldmask.WithCause(err)
			}
			if err := cmd.ScanProto(updated); err != nil {
				return err
			}
			updated, err = applyApplicationAPIKeyFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}
		}

		if err := updated.ValidateFields(sets...); err != nil {
			return err
		}

		_, err = ttnredis.SetProto(tx, uk, updated, 0)
		if err != nil {
			return err
		}
		pb, err = applyApplicationAPIKeyFieldMask(nil, updated, gets...)
		return err
	}, uk)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
