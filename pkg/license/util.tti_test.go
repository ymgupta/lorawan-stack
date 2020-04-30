// Copyright © 2019 The Things Industries B.V.

package license_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	. "go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func TestCheckValidity(t *testing.T) {
	a := assertions.New(t)

	a.So(CheckValidity(&ttipb.License{
		ValidFrom: time.Now().Add(time.Hour),
	}), should.NotBeNil)

	a.So(CheckValidity(&ttipb.License{
		ValidUntil: time.Now().Add(-1 * time.Hour),
	}), should.NotBeNil)

	a.So(CheckValidity(&ttipb.License{
		MinVersion: "4.0.0",
	}), should.NotBeNil)

	a.So(CheckValidity(&ttipb.License{
		MaxVersion: "2.0.0",
	}), should.NotBeNil)
}

func TestCheckLimitedFunctionality(t *testing.T) {
	a := assertions.New(t)

	a.So(CheckLimitedFunctionality(&ttipb.License{
		ValidUntil: time.Now().Add(time.Hour),
	}), should.BeNil)

	a.So(CheckLimitedFunctionality(&ttipb.License{
		ValidUntil: time.Now().Add(time.Hour),
		LimitFor:   2 * time.Hour,
	}), should.NotBeNil)

	a.So(CheckLimitedFunctionality(&ttipb.License{
		ValidUntil: time.Now().Add(-1 * time.Hour),
	}), should.NotBeNil)
}
