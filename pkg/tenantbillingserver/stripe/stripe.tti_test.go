// Copyright © 2019 The Things Industries B.V.

package stripe_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	stripego "github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/tenantbillingserver/stripe"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// The test data is obtained from the webhooks forwarded by the Stripe CLI tool
// (https://github.com/stripe/stripe-cli). In order to regenerate it, it is necessary
// to create a product in Stripe, add two billing plans to it (one metered and recurring),
// and then create a customer with a subscription to each of the plans.
// The test data is tied to the Stripe API version, which itself is tied to the stripe-go library
// for the backend. As such, it should be regenerated only after updating the stripe-go library.

const (
	tenantID     = "testclient1"
	tenantName   = "TestClient1"
	stripeAPIKey = "valid-api-key"

	// The IDs can be obtained from both the dumped test data, and from the Stripe web UI.
	recurringPlanID           = "plan_G6CmtjyI1lsDhn"
	meteredPlanID             = "plan_FyH3GcT0Q0s5uc"
	meteredSubscriptionItemID = "si_G7O9VTXtw5gGVU"
	customerID                = "cus_G7JW3JDioJTVcq"
)

func TestRecurringPlan(t *testing.T) {
	withStripe(t, func(t *testing.T, s *stripe.Stripe, tnt *mockTenantClient, strp *mockStripeBackend) {
		a := assertions.New(t)
		ctx := test.Context()

		strp.CallMock = func(method, path, key string, params stripego.ParamsContainer, v interface{}) error {
			switch method {
			case http.MethodGet:
			default:
				t.Fatalf("Unexpected method received %v", method)
			}
			switch {
			case strings.HasPrefix(path, "/v1/customers/"+customerID):
				return json.Unmarshal(recurringCustomerData, v)
			default:
				t.Fatalf("Unexpected call received on path %v", path)
			}
			return nil
		}

		url := fmt.Sprintf("http://127.0.0.1:8099/api/v3/tbs/stripe")
		client := createHTTPClient()

		// Create the subscription, without the tenant ID.
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(recurringSubscriptionCreatedEventData))
		a.So(err, should.BeNil)

		resp, err := client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.BeEmpty)

		// Update the subscription, now containing the tenant ID.
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(recurringSubscriptionUpdatedEventData))
		a.So(err, should.BeNil)

		resp, err = client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.HaveLength, 1)
		a.So(tnt.tenants, should.ContainKey, tenantID)
		a.So(tnt.tenants[tenantID].Name, should.Equal, tenantName)
		a.So(tnt.tenants[tenantID].State, should.Equal, ttnpb.STATE_APPROVED)
		a.So(tnt.tenants[tenantID].MaxApplications.Value, should.Equal, 123)
		a.So(tnt.tenants[tenantID].MaxClients.Value, should.Equal, 124)
		a.So(tnt.tenants[tenantID].MaxEndDevices.Value, should.Equal, 125)
		a.So(tnt.tenants[tenantID].MaxGateways.Value, should.Equal, 126)
		a.So(tnt.tenants[tenantID].MaxOrganizations.Value, should.Equal, 127)
		a.So(tnt.tenants[tenantID].MaxUsers.Value, should.Equal, 128)

		// No-op on recurring plan.
		// Since the Call mock doesn't handle usage records the error will occur there.
		err = s.Report(ctx, tnt.tenants[tenantID], &ttipb.TenantRegistryTotals{})
		a.So(err, should.BeNil)

		// Cancel the subscription.
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(recurringSubscriptionDeletedEventData))
		a.So(err, should.BeNil)

		resp, err = client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.HaveLength, 1)
		a.So(tnt.tenants, should.ContainKey, tenantID)
		a.So(tnt.tenants[tenantID].State, should.Equal, ttnpb.STATE_SUSPENDED)

		// No-op on suspended tenants.
		err = s.Report(ctx, tnt.tenants[tenantID], &ttipb.TenantRegistryTotals{})
		a.So(err, should.BeNil)
	})
}

func TestMeteredPlan(t *testing.T) {
	withStripe(t, func(t *testing.T, s *stripe.Stripe, tnt *mockTenantClient, strp *mockStripeBackend) {
		a := assertions.New(t)
		ctx := test.Context()

		strp.CallMock = func(method, path, key string, params stripego.ParamsContainer, v interface{}) error {
			switch method {
			case http.MethodGet:
				switch {
				case strings.HasPrefix(path, "/v1/customers/"+customerID):
					return json.Unmarshal(meteredCustomerData, v)
				default:
					t.Fatalf("Unexpected GET call received on path %v", path)
				}
			case http.MethodPost:
				switch {
				case strings.HasPrefix(path, fmt.Sprintf("/v1/subscription_items/%v/usage_records", meteredSubscriptionItemID)):
					return json.Unmarshal(meteredUsageRecordData, v)
				default:
					t.Fatalf("Unexpected POST call received on path %v", path)
				}
			default:
				t.Fatalf("Unexpected method received %v", method)
			}
			return nil
		}

		url := fmt.Sprintf("http://127.0.0.1:8099/api/v3/tbs/stripe")
		client := createHTTPClient()

		// Create the subscription, without the tenant ID.
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(meteredSubscriptionCreatedEventData))
		a.So(err, should.BeNil)

		resp, err := client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.BeEmpty)

		// Update the subscription, now containing the tenant ID.
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(meteredSubscriptionUpdatedEventData))
		a.So(err, should.BeNil)

		resp, err = client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.HaveLength, 1)
		a.So(tnt.tenants, should.ContainKey, tenantID)
		a.So(tnt.tenants[tenantID].Name, should.Equal, tenantName)
		a.So(tnt.tenants[tenantID].State, should.Equal, ttnpb.STATE_APPROVED)
		a.So(tnt.tenants[tenantID].MaxApplications, should.BeNil)
		a.So(tnt.tenants[tenantID].MaxClients, should.BeNil)
		a.So(tnt.tenants[tenantID].MaxEndDevices, should.BeNil)
		a.So(tnt.tenants[tenantID].MaxGateways, should.BeNil)
		a.So(tnt.tenants[tenantID].MaxOrganizations, should.BeNil)
		a.So(tnt.tenants[tenantID].MaxUsers, should.BeNil)

		// Simulate a report coming from the license package.
		err = s.Report(ctx, tnt.tenants[tenantID], &ttipb.TenantRegistryTotals{
			EndDevices: 345,
		})
		a.So(err, should.BeNil)

		// Cancel the subscription.
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(meteredSubscriptionDeletedEventData))
		a.So(err, should.BeNil)

		resp, err = client.Do(req)
		a.So(err, should.BeNil)
		a.So(resp, should.NotBeNil)

		a.So(tnt.tenants, should.HaveLength, 1)
		a.So(tnt.tenants, should.ContainKey, tenantID)
		a.So(tnt.tenants[tenantID].State, should.Equal, ttnpb.STATE_SUSPENDED)

		// No-op on suspended tenants.
		err = s.Report(ctx, tnt.tenants[tenantID], &ttipb.TenantRegistryTotals{})
		a.So(err, should.BeNil)
	})
}

func withStripe(t *testing.T, testFunc func(*testing.T, *stripe.Stripe, *mockTenantClient, *mockStripeBackend)) {
	ctx := test.Context()
	a := assertions.New(t)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			HTTP: config.HTTP{
				Listen: ":8099",
			},
		},
	})

	config := stripe.Config{
		Enable:                  true,
		APIKey:                  stripeAPIKey,
		SkipSignatureValidation: true,
		RecurringPlanIDs:        []string{recurringPlanID},
		MeteredPlanIDs:          []string{meteredPlanID},
	}

	tnt := &mockTenantClient{}
	tnt.tenants = make(map[string]*ttipb.Tenant)
	backend := &mockStripeBackend{}
	strp := createStripeMock(ctx, stripeAPIKey, backend)

	s, err := config.New(ctx, c, stripe.WithTenantRegistryClient(tnt), stripe.WithStripeAPIClient(strp))
	a.So(s, should.NotBeNil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterWeb(s)

	componenttest.StartComponent(t, c)
	defer c.Close()

	testFunc(t, s, tnt, backend)
}

var (
	recurringSubscriptionCreatedEventData []byte
	recurringSubscriptionUpdatedEventData []byte
	recurringSubscriptionDeletedEventData []byte
	recurringCustomerData                 []byte

	meteredSubscriptionCreatedEventData []byte
	meteredSubscriptionUpdatedEventData []byte
	meteredSubscriptionDeletedEventData []byte
	meteredCustomerData                 []byte
	meteredUsageRecordData              []byte
)

func init() {
	for _, f := range []struct {
		destination *[]byte
		file        string
	}{
		{
			destination: &recurringCustomerData,
			file:        "testdata/recurring/customer.json",
		},
		{
			destination: &recurringSubscriptionCreatedEventData,
			file:        "testdata/recurring/subscription-created.json",
		},
		{
			destination: &recurringSubscriptionDeletedEventData,
			file:        "testdata/recurring/subscription-deleted.json",
		},
		{
			destination: &recurringSubscriptionUpdatedEventData,
			file:        "testdata/recurring/subscription-updated.json",
		},
		{
			destination: &meteredCustomerData,
			file:        "testdata/metered/customer.json",
		},
		{
			destination: &meteredSubscriptionCreatedEventData,
			file:        "testdata/metered/subscription-created.json",
		},
		{
			destination: &meteredSubscriptionDeletedEventData,
			file:        "testdata/metered/subscription-deleted.json",
		},
		{
			destination: &meteredSubscriptionUpdatedEventData,
			file:        "testdata/metered/subscription-updated.json",
		},
		{
			destination: &meteredUsageRecordData,
			file:        "testdata/metered/usage-record.json",
		},
	} {
		var err error
		*f.destination, err = ioutil.ReadFile(f.file)
		if err != nil {
			panic(err)
		}
	}
}
