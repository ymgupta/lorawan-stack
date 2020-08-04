// Copyright © 2020 The Things Industries B.V.

package oidc

type claimsToken struct {
	ExpiresAt uint64 `json:"exp"`
	IssuedAt  uint64 `json:"iat"`

	IssuerIdentifier string `json:"iss"`
	Subject          string `json:"sub"`

	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`

	FamilyName   string `json:"family_name"`
	GivenName    string `json:"given_name"`
	Name         string `json:"name"`
	HostedDomain string `json:"hd"`
	Locale       string `json:"locale"`

	PictureURL string `json:"picture"`
	ProfileURL string `json:"profile"`
}
