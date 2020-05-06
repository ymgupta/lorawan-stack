// Copyright © 2019 The Things Industries B.V.

package awsmetrics_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/license/awsmetrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestComputeUsage(t *testing.T) {
	for _, tc := range []struct {
		deviceCount       uint64
		expectedDimension string
		expectedQuantity  int64
	}{
		{
			deviceCount:       500,
			expectedDimension: "1000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       1500,
			expectedDimension: "2000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       2500,
			expectedDimension: "4000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       4500,
			expectedDimension: "6500devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       7500,
			expectedDimension: "10000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       12500,
			expectedDimension: "15000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       17500,
			expectedDimension: "20000devices",
			expectedQuantity:  1,
		},
		{
			deviceCount:       22000,
			expectedDimension: "Up20000devices",
			expectedQuantity:  220,
		},
		{
			deviceCount:       22513,
			expectedDimension: "Up20000devices",
			expectedQuantity:  226,
		},
	} {
		t.Run(fmt.Sprintf("%dEndDevices", tc.deviceCount), func(t *testing.T) {
			a := assertions.New(t)

			dimension, quantity := ComputeUsage(&ttipb.MeteringData{
				Tenants: []*ttipb.MeteringData_TenantMeteringData{
					{
						Totals: &ttipb.TenantRegistryTotals{
							EndDevices: tc.deviceCount,
						},
					},
				},
			})
			if !a.So(dimension, should.NotBeNil) || !a.So(quantity, should.NotBeNil) {
				t.FailNow()
			}
			a.So(*dimension, should.Equal, tc.expectedDimension)
			a.So(*quantity, should.Equal, tc.expectedQuantity)
		})
	}
}
