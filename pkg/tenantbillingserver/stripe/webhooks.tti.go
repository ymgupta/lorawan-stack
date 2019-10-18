// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"encoding/json"
	"io/ioutil"

	echo "github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"go.thethings.network/lorawan-stack/pkg/log"
)

const (
	customerSubscriptionCreated = "customer.subscription.created"
	customerSubscriptionUpdated = "customer.subscription.updated"
	customerSubscriptionDeleted = "customer.subscription.deleted"

	stripeSignatureHeader = "Stripe-Signature"
)

func (s *Stripe) handleWebhook(c echo.Context) error {
	ctx := s.component.FillContext(c.Request().Context())
	logger := log.FromContext(ctx)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	event, err := webhook.ConstructEvent(body, c.Request().Header.Get(stripeSignatureHeader), s.config.EndpointSecretKey)
	if err != nil {
		logger.WithError(err).Warn("Webhook signature validation failed")
		return err
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

	switch sub.Status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return s.createTenant(ctx, sub)
	case stripe.SubscriptionStatusIncomplete, stripe.SubscriptionStatusIncompleteExpired:
		return nil
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid:
		return s.deleteTenant(ctx, sub)
	default:
		panic("unreachable")
	}
}
