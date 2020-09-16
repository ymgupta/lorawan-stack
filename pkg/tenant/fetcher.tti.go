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

type combinedFieldsFetcher struct {
	Fetcher
	mu         sync.Mutex
	fieldPaths map[string]struct{}
}

// NewCombinedFieldsFetcher returns a fetcher that combines fields of subsequent fetches.
func NewCombinedFieldsFetcher(fetcher Fetcher) Fetcher {
	return &combinedFieldsFetcher{
		Fetcher:    fetcher,
		fieldPaths: make(map[string]struct{}),
	}
}

func (f *combinedFieldsFetcher) FetchTenant(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
	f.mu.Lock()
	for _, path := range fieldPaths {
		if _, exists := f.fieldPaths[path]; exists {
			continue
		}
		f.fieldPaths[path] = struct{}{}
	}
	fetchPaths := make([]string, 0, len(f.fieldPaths))
	for path := range f.fieldPaths {
		fetchPaths = append(fetchPaths, path)
	}
	f.mu.Unlock()
	sort.Strings(fetchPaths)
	return f.Fetcher.FetchTenant(ctx, ids, fetchPaths...)
}
