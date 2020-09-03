// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"fmt"
	"regexp"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/random"
)

func removeDuplicateRune(s string, d rune) string {
	var buff strings.Builder
	var last rune
	for i, r := range s {
		if r != last || r != d || i == 0 {
			buff.WriteRune(r)
			last = r
		}
	}
	return buff.String()
}

var alphanumericRegexp = regexp.MustCompile("[^a-zA-Z0-9]+")

func cleanUsername(username string) string {
	lower := strings.ToLower(username)
	alphanumeric := alphanumericRegexp.ReplaceAllString(lower, "-")
	alphanumeric = strings.Trim(alphanumeric, "-")
	unique := removeDuplicateRune(alphanumeric, '-')
	n := 36
	if len(unique) < n {
		n = len(unique)
	}
	return unique[:n]
}

func usernames(provider string, ct *claimsToken) []string {
	// Username derived from the given and family name.
	nameBase := fmt.Sprintf("%s-%s", ct.GivenName, ct.FamilyName)
	nameProviderBase := fmt.Sprintf("%s-%s", nameBase, provider)
	// Username derived from the root of the email address.
	emailBase := strings.SplitN(ct.Email, "@", 2)[0]
	emailProviderBase := fmt.Sprintf("%s-%s", emailBase, provider)

	// Usernames derived from the bases with a random suffix.
	nameBased := make([]string, 0, 5)
	nameProviderBased := make([]string, 0, 5)
	emailBased := make([]string, 0, 5)
	emailProviderBased := make([]string, 0, 5)
	for i := 2; i < 7; i++ {
		nameBased = append(nameBased, fmt.Sprintf("%s-%s", nameBase, random.String(i)))
		nameProviderBased = append(nameProviderBased, fmt.Sprintf("%s-%s", nameProviderBase, random.String(i)))
		emailBased = append(emailBased, fmt.Sprintf("%s-%s", emailBase, random.String(i)))
		emailProviderBased = append(emailProviderBased, fmt.Sprintf("%s-%s", emailProviderBase, random.String(i)))
	}

	usernames := []string{}
	if ct.PreferredUsername != "" {
		usernames = append(usernames, ct.PreferredUsername)
	}
	usernames = append(usernames, emailBase)
	usernames = append(usernames, emailBased...)
	usernames = append(usernames, nameBase)
	usernames = append(usernames, nameBased...)
	usernames = append(usernames, emailProviderBase)
	usernames = append(usernames, emailProviderBased...)
	usernames = append(usernames, nameProviderBase)
	usernames = append(usernames, nameProviderBased...)

	cleanUsers := make([]string, 0, len(usernames))
	for _, v := range usernames {
		if clean := cleanUsername(v); clean != "" {
			cleanUsers = append(cleanUsers, clean)
		}
	}
	return cleanUsers
}
