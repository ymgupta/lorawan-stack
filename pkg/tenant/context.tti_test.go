// Copyright © 2019 The Things Industries B.V.

package tenant_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	. "go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestFromContext(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assertions.New(t).So(FromContext(context.Background()).IsZero(), should.BeTrue)
	})

	t.Run("Test Context", func(t *testing.T) {
		assertions.New(t).So(FromContext(test.Context()).TenantID, should.Equal, "foo-tenant")
	})

	t.Run("Context With Tenant", func(t *testing.T) {
		assertions.New(t).So(FromContext(NewContext(context.Background(), ttipb.TenantIdentifiers{TenantID: "foo-bar"})).TenantID, should.Equal, "foo-bar")
	})
}
