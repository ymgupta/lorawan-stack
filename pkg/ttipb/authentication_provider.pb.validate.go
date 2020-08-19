// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttipb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogo/protobuf/types"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = types.DynamicAny{}
)

// define the regex for a UUID once up-front
var _authentication_provider_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// ValidateFields checks the field values on AuthenticationProviderIdentifiers
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *AuthenticationProviderIdentifiers) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = AuthenticationProviderIdentifiersFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "provider_id":

			if utf8.RuneCountInString(m.GetProviderID()) > 36 {
				return AuthenticationProviderIdentifiersValidationError{
					field:  "provider_id",
					reason: "value length must be at most 36 runes",
				}
			}

			if !_AuthenticationProviderIdentifiers_ProviderID_Pattern.MatchString(m.GetProviderID()) {
				return AuthenticationProviderIdentifiersValidationError{
					field:  "provider_id",
					reason: "value does not match regex pattern \"^[a-z0-9](?:[-]?[a-z0-9]){2,}$\"",
				}
			}

		default:
			return AuthenticationProviderIdentifiersValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

// AuthenticationProviderIdentifiersValidationError is the validation error
// returned by AuthenticationProviderIdentifiers.ValidateFields if the
// designated constraints aren't met.
type AuthenticationProviderIdentifiersValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AuthenticationProviderIdentifiersValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AuthenticationProviderIdentifiersValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AuthenticationProviderIdentifiersValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AuthenticationProviderIdentifiersValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AuthenticationProviderIdentifiersValidationError) ErrorName() string {
	return "AuthenticationProviderIdentifiersValidationError"
}

// Error satisfies the builtin error interface
func (e AuthenticationProviderIdentifiersValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAuthenticationProviderIdentifiers.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AuthenticationProviderIdentifiersValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AuthenticationProviderIdentifiersValidationError{}

var _AuthenticationProviderIdentifiers_ProviderID_Pattern = regexp.MustCompile("^[a-z0-9](?:[-]?[a-z0-9]){2,}$")

// ValidateFields checks the field values on AuthenticationProvider with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *AuthenticationProvider) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = AuthenticationProviderFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "ids":

			if v, ok := interface{}(&m.AuthenticationProviderIdentifiers).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return AuthenticationProviderValidationError{
						field:  "ids",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		case "created_at":

			if v, ok := interface{}(&m.CreatedAt).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return AuthenticationProviderValidationError{
						field:  "created_at",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		case "updated_at":

			if v, ok := interface{}(&m.UpdatedAt).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return AuthenticationProviderValidationError{
						field:  "updated_at",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		case "name":

			if utf8.RuneCountInString(m.GetName()) > 50 {
				return AuthenticationProviderValidationError{
					field:  "name",
					reason: "value length must be at most 50 runes",
				}
			}

		case "allow_registrations":
			// no validation rules for AllowRegistrations
		case "configuration":

			if v, ok := interface{}(m.GetConfiguration()).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return AuthenticationProviderValidationError{
						field:  "configuration",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		default:
			return AuthenticationProviderValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

// AuthenticationProviderValidationError is the validation error returned by
// AuthenticationProvider.ValidateFields if the designated constraints aren't met.
type AuthenticationProviderValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AuthenticationProviderValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AuthenticationProviderValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AuthenticationProviderValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AuthenticationProviderValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AuthenticationProviderValidationError) ErrorName() string {
	return "AuthenticationProviderValidationError"
}

// Error satisfies the builtin error interface
func (e AuthenticationProviderValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAuthenticationProvider.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AuthenticationProviderValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AuthenticationProviderValidationError{}

// ValidateFields checks the field values on AuthenticationProvider_OIDC with
// the rules defined in the proto definition for this message. If any rules
// are violated, an error is returned.
func (m *AuthenticationProvider_OIDC) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = AuthenticationProvider_OIDCFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "client_id":
			// no validation rules for ClientID
		case "client_secret":
			// no validation rules for ClientSecret
		case "provider_url":

			if uri, err := url.Parse(m.GetProviderURL()); err != nil {
				return AuthenticationProvider_OIDCValidationError{
					field:  "provider_url",
					reason: "value must be a valid URI",
					cause:  err,
				}
			} else if !uri.IsAbs() {
				return AuthenticationProvider_OIDCValidationError{
					field:  "provider_url",
					reason: "value must be absolute",
				}
			}

		default:
			return AuthenticationProvider_OIDCValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

// AuthenticationProvider_OIDCValidationError is the validation error returned
// by AuthenticationProvider_OIDC.ValidateFields if the designated constraints
// aren't met.
type AuthenticationProvider_OIDCValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AuthenticationProvider_OIDCValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AuthenticationProvider_OIDCValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AuthenticationProvider_OIDCValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AuthenticationProvider_OIDCValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AuthenticationProvider_OIDCValidationError) ErrorName() string {
	return "AuthenticationProvider_OIDCValidationError"
}

// Error satisfies the builtin error interface
func (e AuthenticationProvider_OIDCValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAuthenticationProvider_OIDC.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AuthenticationProvider_OIDCValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AuthenticationProvider_OIDCValidationError{}

// ValidateFields checks the field values on
// AuthenticationProvider_Configuration with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *AuthenticationProvider_Configuration) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = AuthenticationProvider_ConfigurationFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "provider":
			if m.Provider == nil {
				return AuthenticationProvider_ConfigurationValidationError{
					field:  "provider",
					reason: "value is required",
				}
			}
			if len(subs) == 0 {
				subs = []string{
					"oidc",
				}
			}
			for name, subs := range _processPaths(subs) {
				_ = subs
				switch name {
				case "oidc":
					w, ok := m.Provider.(*AuthenticationProvider_Configuration_OIDC)
					if !ok || w == nil {
						continue
					}

					if v, ok := interface{}(m.GetOIDC()).(interface{ ValidateFields(...string) error }); ok {
						if err := v.ValidateFields(subs...); err != nil {
							return AuthenticationProvider_ConfigurationValidationError{
								field:  "oidc",
								reason: "embedded message failed validation",
								cause:  err,
							}
						}
					}

				}
			}
		default:
			return AuthenticationProvider_ConfigurationValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

// AuthenticationProvider_ConfigurationValidationError is the validation error
// returned by AuthenticationProvider_Configuration.ValidateFields if the
// designated constraints aren't met.
type AuthenticationProvider_ConfigurationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AuthenticationProvider_ConfigurationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AuthenticationProvider_ConfigurationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AuthenticationProvider_ConfigurationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AuthenticationProvider_ConfigurationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AuthenticationProvider_ConfigurationValidationError) ErrorName() string {
	return "AuthenticationProvider_ConfigurationValidationError"
}

// Error satisfies the builtin error interface
func (e AuthenticationProvider_ConfigurationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAuthenticationProvider_Configuration.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AuthenticationProvider_ConfigurationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AuthenticationProvider_ConfigurationValidationError{}
