// Copyright © 2019 The Things Industries B.V.

package aws_test

import (
	"crypto/x509"
	"encoding/hex"
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto/aws"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// TestKeyVault tests the AWS key vault implementation.
//
// This test requires AWS_REGION, AWS_SECRET_KEY_ID and AWS_SECRET_ACCESS_KEY to be set in the environment.
//
// For testing wrapping, in AWS Secrets Manager, create a secret with ID `testing/kek` with a field `value` set to `AAECAwQFBgcICQoLDA0ODw==`.
// You can override the secret ID by setting TEST_AWS_KEYVAULT_KEK_SECRET_ID.
// The key vault needs secretsmanager:GetSecretValue for key wrapping.
//
// For testing certificates, in AWS Secrets Manager, create a secret with ID `testing/certificate` with a field
// `certificate` and `key` set to a PEM encoded certificate and private key.
// You can override the secret ID by setting TEST_AWS_KEYVAULT_CERTIFICATE_SECRET_ID.
// The key vault needs secretsmanager:GetSecretValue for getting and exporting the certificate.
func TestKeyVault(t *testing.T) {
	a := assertions.New(t)

	region := os.Getenv("AWS_REGION")
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if region == "" || accessKeyID == "" || secretAccessKey == "" {
		t.Skip("Missing AWS credentials")
	}

	kv, err := aws.NewKeyVault(region, "")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	t.Run("Wrapping", func(t *testing.T) {
		a := assertions.New(t)

		a.So(kv.AsKEKLabel(test.Context(), "foo-tenant.example.com:8884"), should.Equal, "as/example.com")
		a.So(kv.AsKEKLabel(test.Context(), "[::1]:8884"), should.Equal, "as/__1")
		a.So(kv.NsKEKLabel(test.Context(), &types.NetID{0x0, 0x0, 0x13}, "foo-tenant.example.com:8884"), should.Equal, "ns/000013/example.com")

		plaintext, _ := hex.DecodeString("00112233445566778899AABBCCDDEEFF")
		ciphertext, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

		kekLabel := os.Getenv("TEST_AWS_KEYVAULT_KEK_SECRET_ID")
		if kekLabel == "" {
			kekLabel = "testing/kek"
		}

		wrapped, err := kv.Wrap(test.Context(), plaintext, kekLabel)
		if !a.So(err, should.BeNil) || !a.So(wrapped, should.Resemble, ciphertext) {
			t.FailNow()
		}

		unwrapped, err := kv.Unwrap(test.Context(), wrapped, kekLabel)
		if !a.So(err, should.BeNil) || !a.So(unwrapped, should.Resemble, plaintext) {
			t.FailNow()
		}
	})

	t.Run("Certificate", func(t *testing.T) {
		a := assertions.New(t)

		id := os.Getenv("TEST_AWS_KEYVAULT_CERTIFICATE_SECRET_ID")
		if id == "" {
			id = "testing/certificate"
		}

		cert, err := kv.GetCertificate(test.Context(), id)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		t.Logf("Got certificate with CN: %s", cert.Subject.CommonName)

		keyPair, err := kv.ExportCertificate(test.Context(), id)
		if !a.So(err, should.BeNil) || !a.So(len(keyPair.Certificate), should.BeGreaterThanOrEqualTo, 1) {
			t.FailNow()
		}

		certFromPair, err := x509.ParseCertificate(keyPair.Certificate[0])
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(certFromPair.Subject, should.Resemble, cert.Subject)
	})
}
