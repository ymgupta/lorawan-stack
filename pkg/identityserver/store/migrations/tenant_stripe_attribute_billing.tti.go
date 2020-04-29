// Copyright © 2020 The Things Industries B.V.

package migrations

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

// TenantStripeAttributeBilling migrates the Stripe tenant billing information
// from the attributes to the specific billing field. Does not change the attributes.
type TenantStripeAttributeBilling struct{}

func (TenantStripeAttributeBilling) Name() string {
	return "tenant_stripe_attribute_billing"
}

func (TenantStripeAttributeBilling) columns() []string {
	return []string{"id", "created_at", "updated_at", "tenant_id", "billing"}
}

const (
	managerAttribute            = "managed-by"
	managerAttributeValue       = "stripe"
	customerIDAttribute         = "stripe-customer-id"
	planIDAttribute             = "stripe-plan-id"
	subscriptionIDAttribute     = "stripe-subscription-id"
	subscriptionItemIDAttribute = "stripe-subscription-item-id"
)

func (m TenantStripeAttributeBilling) Apply(ctx context.Context, db *gorm.DB) error {
	var models []store.Tenant
	err := db.Model(store.Tenant{}).Select(m.columns()).Preload("Attributes").Find(&models).Error
	if err != nil {
		return err
	}
	logger := log.FromContext(ctx)
	for _, model := range models {
		logger := logger.WithField("tenant_id", model.TenantID)
		billing := &ttipb.Billing{}
		if len(model.Billing.RawMessage) > 0 {
			if err := jsonpb.TTN().Unmarshal(model.Billing.RawMessage, billing); err != nil {
				return err
			}
			if billing.Provider != nil {
				continue
			}
		}

		attributes := make(map[string]string)
		for _, attribute := range model.Attributes {
			attributes[attribute.Key] = attribute.Value
		}
		manager, ok := attributes[managerAttribute]
		if !ok || manager != managerAttributeValue {
			continue
		}
		customerID, ok := attributes[customerIDAttribute]
		if !ok {
			logger.Warn("Missing Stripe customer ID")
			continue
		}
		planID, ok := attributes[planIDAttribute]
		if !ok {
			logger.Warn("Missing Stripe plan ID")
			continue
		}
		subscriptionID, ok := attributes[subscriptionIDAttribute]
		if !ok {
			logger.Warn("Missing Stripe subscription ID")
			continue
		}
		subscriptionItemID, ok := attributes[subscriptionItemIDAttribute]
		if !ok {
			logger.Warn("Missing Stripe subscription item ID")
			continue
		}

		billing.Provider = &ttipb.Billing_Stripe_{
			Stripe: &ttipb.Billing_Stripe{
				CustomerID:         customerID,
				PlanID:             planID,
				SubscriptionID:     subscriptionID,
				SubscriptionItemID: subscriptionItemID,
			},
		}
		model.Billing.RawMessage, err = jsonpb.TTN().Marshal(billing)
		if err != nil {
			return err
		}
		if err = db.Select(m.columns()).Save(&model).Error; err != nil {
			return err
		}
	}
	return nil
}
func (m TenantStripeAttributeBilling) Rollback(ctx context.Context, db *gorm.DB) error {
	// No-op since attributes are not deleted during the migration.
	return nil
}
