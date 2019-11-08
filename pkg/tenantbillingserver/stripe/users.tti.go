// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"time"

	"github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	defaultAdminUserID           = "admin"
	adminUserIDStripeAttribute   = "admin-user"
	adminPasswordStripeAttribute = "admin-password"
)

func (s *Stripe) generateInitialUser(ctx context.Context, sub *stripe.Subscription, cust *stripe.Customer) (*ttnpb.User, error) {
	password, ok := sub.Metadata[adminPasswordStripeAttribute]
	if !ok {
		password = random.String(256)
	}
	hashedPassword, err := auth.Hash(ctx, password)
	if err != nil {
		return nil, err
	}

	userID, ok := sub.Metadata[adminUserIDStripeAttribute]
	if !ok {
		userID = defaultAdminUserID
	}

	now := time.Now()
	return &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{
			UserID: userID,
		},
		PrimaryEmailAddress:            cust.Email,
		PrimaryEmailAddressValidatedAt: &now,
		State:                          ttnpb.STATE_APPROVED,
		Password:                       hashedPassword,
		PasswordUpdatedAt:              &now,
		Admin:                          true,
	}, nil
}
