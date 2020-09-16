// Copyright © 2019 The Things Industries B.V.

package tenant

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"golang.org/x/sync/singleflight"
)

// Fetcher is the interface for tenant fetching.
type Fetcher interface {
	FetchTenant(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error)
}

// FetcherFunc is a func for tenant fetching.
type FetcherFunc func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error)

// FetchTenant satisfies the Fetcher interface.
func (f FetcherFunc) FetchTenant(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
	return f(ctx, ids, fieldPaths...)
}

type singleFlightFetcher struct {
	Fetcher
	singleflight singleflight.Group
}

// NewSingleFlightFetcher returns a fetcher that de-duplicates concurrent fetches for the same arguments.
func NewSingleFlightFetcher(fetcher Fetcher) Fetcher {
	return &singleFlightFetcher{Fetcher: fetcher}
}

func (f *singleFlightFetcher) FetchTenant(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
	fieldPaths = normalizeFieldPaths(fieldPaths)
	key := fmt.Sprintf("%s:%s", ids.IDString(), strings.Join(fieldPaths, ","))
	res, err, _ := f.singleflight.Do(key, func() (interface{}, error) {
		return f.Fetcher.FetchTenant(ctx, ids, fieldPaths...)
	})
	if err != nil {
		return nil, err
	}
	return res.(*ttipb.Tenant), nil
}

type fetcherKeyType struct{}

var fetcherKey fetcherKeyType

// NewContextWithFetcher returns a new context with the given tenant fetcher.
func NewContextWithFetcher(ctx context.Context, fetcher Fetcher) context.Context {
	return context.WithValue(ctx, fetcherKey, fetcher)
}

// FetcherFromContext returns the tenant fetcher from the context.
func FetcherFromContext(ctx context.Context) (Fetcher, bool) {
	if fetcher, ok := ctx.Value(fetcherKey).(Fetcher); ok {
		return fetcher, true
	}
	return nil, false
}

func normalizeFieldPaths(fieldPaths []string) []string {
	fieldPathsCopy := make([]string, len(fieldPaths))
	copy(fieldPathsCopy, fieldPaths)
	sort.Strings(fieldPathsCopy)
	return fieldPathsCopy
}

type cachedTenant struct {
	tenant  *ttipb.Tenant
	err     error
	expires time.Time
}

type cachedFetcher struct {
	Fetcher
	ttlFunc func(error) time.Duration
	mu      sync.Mutex
	cache   map[string]*cachedTenant
}

var timeNow = time.Now

func (c *cachedFetcher) FetchTenant(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
	fieldPaths = normalizeFieldPaths(fieldPaths)
	key := fmt.Sprintf("%s:%s", ids.IDString(), strings.Join(fieldPaths, ","))
	c.mu.Lock()
	cached, wasCached := c.cache[key]
	if !wasCached {
		cached = &cachedTenant{}
		c.cache[key] = cached
	}
	cacheValid := timeNow().Before(cached.expires)
	c.mu.Unlock()
	if !cacheValid {
		tenant, err := c.Fetcher.FetchTenant(ctx, ids, fieldPaths...)
		if err == nil {
			cached.tenant, cached.err = tenant, err
		} else {
			cached.err = err // keep the old tenant.
		}
		cached.expires = timeNow().Add(c.ttlFunc(cached.err))
	}
	return cached.tenant, cached.err
}

// Expire expires all cached tenants at the given time.
// Typically only used in tests.
func (c *cachedFetcher) Expire(t time.Time) {
	c.mu.Lock()
	for _, cached := range c.cache {
		cached.expires = t
	}
	c.mu.Unlock()
}

// StaticTTL returns a TTL func that always returns the given TTL.
func StaticTTL(ttl time.Duration) func(error) time.Duration {
	return func(error) time.Duration {
		return ttl
	}
}

// NewCachedFetcher wraps the fetcher with a cache.
func NewCachedFetcher(fetcher Fetcher, ttlFunc func(error) time.Duration) Fetcher {
	return &cachedFetcher{
		Fetcher: fetcher,
		ttlFunc: ttlFunc,
		cache:   make(map[string]*cachedTenant),
	}
}

// NewMapFetcher returns a new tenant fetcher that returns tenants from a map.
// The Map Fetcher should typically be used for testing only.
func NewMapFetcher(tenants map[string]*ttipb.Tenant) Fetcher { return mapFetcher(tenants) }

type mapFetcher map[string]*ttipb.Tenant

var errTenantNotFound = errors.DefineNotFound("tenant_not_found", "tenant `{tenant_id}` not found")

func (f mapFetcher) FetchTenant(_ context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
	tenant, ok := f[ids.TenantID]
	if !ok {
		return nil, errTenantNotFound.WithAttributes("tenant_id", ids.TenantID)
	}
	var res ttipb.Tenant
	if len(fieldPaths) == 0 {
		fieldPaths = ttipb.TenantFieldPathsTopLevel
	}
	if err := res.SetFields(tenant, fieldPaths...); err != nil {
		return nil, err
	}
	return &res, nil
}
