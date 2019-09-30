// Copyright © 2019 The Things Industries B.V.

//+build tti

package test

import (
	"testing"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

// NewComponent returns a new Component that can be used for testing.
func NewComponent(t *testing.T, config *component.Config, opts ...component.Option) *component.Component {
	c, err := component.New(test.GetLogger(t), config, opts...)
	if err != nil {
		t.Fatalf("Failed to create component: %v", err)
	}
	return c
}

// StartComponent starts the component for testing.
func StartComponent(t *testing.T, c *component.Component) {
	c.AddContextFiller(test.TenantContextFiller)
	if err := c.Start(); err != nil {
		t.Fatalf("Failed to start component: %v", err)
	}
}
