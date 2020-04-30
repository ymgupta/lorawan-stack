// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/deviceclaimingserver/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func handleAuthorizedApplicationRegistryTest(t *testing.T, reg AuthorizedApplicationRegistry) {
	a := assertions.New(t)
	ctx := test.Context()
	app1 := &ttipb.ApplicationAPIKey{
		ApplicationIDs: ttnpb.ApplicationIdentifiers{
			ApplicationID: "app-1",
		},
		APIKey: "secret1",
	}
	app2 := &ttipb.ApplicationAPIKey{
		ApplicationIDs: ttnpb.ApplicationIdentifiers{
			ApplicationID: "app-2",
		},
		APIKey: "secret2",
	}

	for ids, app := range map[ttnpb.ApplicationIdentifiers]*ttipb.ApplicationAPIKey{
		app1.ApplicationIDs: app1,
		app2.ApplicationIDs: app2,
	} {
		_, err := reg.Get(ctx, ids, ttipb.ApplicationAPIKeyFieldPathsTopLevel)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.FailNow()
		}

		_, err = reg.Set(ctx, ids, nil, func(pb *ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error) {
			if pb != nil {
				t.Fatal("Application already exists")
			}
			return app, ttipb.ApplicationAPIKeyFieldPathsTopLevel, nil
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		pb, err := reg.Get(ctx, ids, ttipb.ApplicationAPIKeyFieldPathsTopLevel)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(pb, should.HaveEmptyDiff, app)
	}

	for _, ids := range []ttnpb.ApplicationIdentifiers{app1.ApplicationIDs, app2.ApplicationIDs} {
		_, err := reg.Set(ctx, ids, nil, func(_ *ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error) {
			return nil, nil, nil
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = reg.Get(ctx, ids, nil)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.FailNow()
		}
	}
}

func TestAuthorizedApplicationRegistry(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"deviceclaimingserver_test",
	}
	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (reg AuthorizedApplicationRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(t testing.TB) (AuthorizedApplicationRegistry, func() error) {
				cl, flush := test.NewRedis(t, namespace[:]...)
				reg := &redis.AuthorizedApplicationRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				reg, closeFn := tc.New(t)
				if closeFn != nil {
					defer func() {
						if err := closeFn(); err != nil {
							t.Errorf("Failed to close registry: %v", err)
						}
					}()
				}
				t.Run("1st run", func(t *testing.T) { handleAuthorizedApplicationRegistryTest(t, reg) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleAuthorizedApplicationRegistryTest(t, reg) })
			})
		}
	}
}
