// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	echo "github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

const (
	customerSubscriptionCreated = "customer.subscription.created"
	customerSubscriptionUpdated = "customer.subscription.updated"
	customerSubscriptionDeleted = "customer.subscription.deleted"

	signatureHeader = "Stripe-Signature"

	correlationIDFormat = "tbs:stripe:%s"
)

func (s *Stripe) handleWebhook(c echo.Context) error {
	ctx := c.Request().Context()
	logger := log.FromContext(ctx)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	defer c.Request().Body.Close()

	var event stripe.Event
	if s.config.EndpointSecretKey == "" {
		err = json.Unmarshal(body, &event)
		if err != nil {
			return err
		}
	} else {
		event, err = webhook.ConstructEvent(body, c.Request().Header.Get(signatureHeader), s.config.EndpointSecretKey)
		if err != nil {
			return err
		}
	}

	switch event.Type {
	case customerSubscriptionCreated, customerSubscriptionUpdated, customerSubscriptionDeleted:
		break
	default:
		logger.WithFields(log.Fields(
			"event_id", event.ID,
			"event_type", event.Type,
		)).Warn("Unexpected event received")
		return nil
	}

	sub := &stripe.Subscription{}
	if err = json.Unmarshal(event.Data.Raw, sub); err != nil {
		return err
	}

	subscriptionItem := s.findSubscriptionItem(sub.Items)
	if subscriptionItem == nil {
		return nil
	}

	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf(correlationIDFormat, subscriptionItem.ID))
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf(correlationIDFormat, subscriptionItem.Plan.ID))
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf(correlationIDFormat, sub.Customer.ID))
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf(correlationIDFormat, sub.ID))
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"subscription_item_id", subscriptionItem.ID,
		"plan_id", subscriptionItem.Plan.ID,
		"customer_id", sub.Customer.ID,
		"subscription_id", sub.ID,
	))

	switch sub.Status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return s.createTenant(ctx, sub)
	case stripe.SubscriptionStatusIncomplete, stripe.SubscriptionStatusIncompleteExpired:
		return nil
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid:
		return s.suspendTenant(ctx, sub)
	default:
		logger.Errorf("Unhandled Stripe subscription status: %s", sub.Status)
		return nil
	}
}

func (s *Stripe) findSubscriptionItem(items *stripe.SubscriptionItemList) *stripe.SubscriptionItem {
	for _, item := range items.Data {
		if _, ok := s.recurringPlanIDs[item.Plan.ID]; ok {
			return item
		}
		if _, ok := s.meteredPlanIDs[item.Plan.ID]; ok {
			return item
		}
	}
	return nil
}
