// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"regexp"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var usernameRegexp = regexp.MustCompile("^[a-z0-9](?:[-]?[a-z0-9]){2,}$")

func TestNormalUsernames(t *testing.T) {
	ct := &claimsToken{
		PreferredUsername: "foo.@#@#bar.!_@@@85~!@",
		GivenName:         "foo",
		FamilyName:        "bar",
		Email:             "foo.bar97++tti@foo.com",
	}
	usernames := usernames("foo-provider", ct)
	a := assertions.New(t)
	for _, username := range usernames {
		a.So(usernameRegexp.MatchString(username), should.BeTrue)
		a.So(len(username), should.BeGreaterThan, 0)
		a.So(len(username), should.BeLessThanOrEqualTo, 36)
	}
}
