// Copyright © 2019 The Things Industries B.V.

package tenant_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSingleFlightTenantFetcher(t *testing.T) {
	var callsToF uint32
	returnFromF := make(chan struct{})
	f := FetcherFunc(func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
		atomic.AddUint32(&callsToF, 1)
		<-returnFromF
		return &ttipb.Tenant{}, nil
	})

	sff := NewSingleFlightFetcher(f)

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			sff.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "name")
			wg.Done()
		}(i)
		runtime.Gosched()
	}

	time.Sleep(5 * test.Delay)

	close(returnFromF)

	wg.Wait()

	assertions.New(t).So(callsToF, should.Equal, 1)
}

func TestFetcherContext(t *testing.T) {
	a := assertions.New(t)

	f := FetcherFunc(func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
		return &ttipb.Tenant{}, nil
	})

	_, ok := FetcherFromContext(context.Background())
	a.So(ok, should.BeFalse)

	ctx := test.Context()

	_, ok = FetcherFromContext(ctx)
	a.So(ok, should.BeTrue) // pkg/util/test injects a fetcher into the test context.

	ctx = NewContextWithFetcher(ctx, f)

	ffc, ok := FetcherFromContext(ctx)
	a.So(ok, should.BeTrue)
	a.So(ffc, should.Equal, f)
}

func TestCachedFetcher(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		a := assertions.New(t)

		var callsToF uint32
		f := FetcherFunc(func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
			callsToF++
			return &ttipb.Tenant{}, nil
		})

		cf := NewCachedFetcher(f, time.Second, time.Second)

		for i := 0; i < 5; i++ {
			_, err := cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "name")
			a.So(err, should.BeNil)
			a.So(callsToF, should.Equal, 1)
		}

		cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "description")
		a.So(callsToF, should.Equal, 2)

		cf.(interface{ Expire(time.Time) }).Expire(time.Now())

		cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "description")
		a.So(callsToF, should.Equal, 3)
	})

	t.Run("Error", func(t *testing.T) {
		a := assertions.New(t)

		var callsToF uint32
		f := FetcherFunc(func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
			callsToF++
			return nil, errors.New("some error")
		})

		cf := NewCachedFetcher(f, time.Second, time.Second)

		for i := 0; i < 5; i++ {
			_, err := cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "name")
			a.So(err, should.NotBeNil)
			a.So(callsToF, should.Equal, 1)
		}

		cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "description")
		a.So(callsToF, should.Equal, 2)

		cf.(interface{ Expire(time.Time) }).Expire(time.Now())

		cf.FetchTenant(test.Context(), &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "description")
		a.So(callsToF, should.Equal, 3)
	})
}

func TestMapFetcher(t *testing.T) {
	a := assertions.New(t)

	fooTenant := &ttipb.Tenant{
		TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "foo-tenant"},
		Name:              "Foo Tenant",
		Description:       "Foo Tenant Description",
	}
	mf := NewMapFetcher(map[string]*ttipb.Tenant{
		"foo-tenant": fooTenant,
	})

	ctx := test.Context()

	res, err := mf.FetchTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo-tenant"})
	a.So(err, should.BeNil)
	a.So(res, should.Resemble, fooTenant)

	res, err = mf.FetchTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo-tenant"}, "name")
	a.So(err, should.BeNil)
	if a.So(res, should.NotBeNil) {
		a.So(res.Name, should.Resemble, fooTenant.Name)
		a.So(res.Description, should.BeEmpty)
	}

	_, err = mf.FetchTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "bar-tenant"}, "name")
	if a.So(err, should.NotBeNil) {
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
}
