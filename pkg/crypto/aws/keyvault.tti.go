// Copyright © 2019 The Things Industries B.V.

package aws

import (
	"context"
	stdcrypto "crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/bluele/gcache"
	"github.com/youmark/pkcs8"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/random"
)

const (
	kekTTL        = time.Minute
	kekErrTTL     = time.Minute
	kekCacheSize  = 1 << 10
	certTTL       = time.Hour
	certErrTTL    = time.Minute
	certCacheSize = 1 << 7
)

type keyVault struct {
	crypto.ComponentKEKLabeler

	secretIDPrefix string

	secrets *secretsmanager.SecretsManager
	certs   *acm.ACM

	kekCache, kekErrCache,
	certCache, certErrCache,
	certExportCache, certExportErrCache gcache.Cache
}

type kekSecret struct {
	Value []byte `json:"value"`
}

// NewKeyVault returns a new AWS key vault.
func NewKeyVault(region, secretIDPrefix string) (crypto.KeyVault, error) {
	awsconfig := aws.NewConfig()
	if region != "" {
		awsconfig = awsconfig.WithRegion(region)
	}
	ses, err := session.NewSession(awsconfig)
	if err != nil {
		return nil, err
	}
	kv := &keyVault{
		ComponentKEKLabeler: &cryptoutil.ComponentPrefixKEKLabeler{
			Separator:     "/",
			ReplaceOldNew: []string{":", "_"},
		},
		secretIDPrefix:     secretIDPrefix,
		secrets:            secretsmanager.New(ses),
		certs:              acm.New(ses),
		kekCache:           gcache.New(kekCacheSize).Expiration(kekTTL).LFU().Build(),
		kekErrCache:        gcache.New(kekCacheSize).Expiration(kekErrTTL).LFU().Build(),
		certCache:          gcache.New(certCacheSize).Expiration(certTTL).LFU().Build(),
		certErrCache:       gcache.New(certCacheSize).Expiration(certErrTTL).LFU().Build(),
		certExportCache:    gcache.New(certCacheSize).Expiration(certTTL).LFU().Build(),
		certExportErrCache: gcache.New(certCacheSize).Expiration(certErrTTL).LFU().Build(),
	}
	return kv, nil
}

var errSecretContent = errors.DefineAborted("secret_content", "invalid secret content in `{id}`")

func (k *keyVault) loadKEK(ctx context.Context, kekLabel string) (kek []byte, err error) {
	id := kekLabel
	if k.secretIDPrefix != "" {
		id = fmt.Sprintf("%s/%s", k.secretIDPrefix, id)
	}
	if v, err := k.kekErrCache.Get(id); err == nil {
		return nil, v.(error)
	}
	if v, err := k.kekCache.Get(id); err == nil {
		return v.([]byte), nil
	}
	defer func() {
		if err != nil {
			k.kekCache.Remove(id)
			k.kekErrCache.Set(id, err)
		} else {
			k.kekCache.Set(id, kek)
			k.kekErrCache.Remove(id)
		}
	}()
	res, err := k.secrets.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	if res.SecretString == nil {
		return nil, errSecretContent.WithAttributes("id", id)
	}
	var secret kekSecret
	if err := json.Unmarshal([]byte(*res.SecretString), &secret); err != nil {
		return nil, errSecretContent.WithCause(err).WithAttributes("id", id)
	}
	return secret.Value, nil
}

func (k *keyVault) Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error) {
	kek, err := k.loadKEK(ctx, kekLabel)
	if err != nil {
		return nil, err
	}
	return crypto.WrapKey(plaintext, kek)
}

func (k *keyVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	kek, err := k.loadKEK(ctx, kekLabel)
	if err != nil {
		return nil, err
	}
	return crypto.UnwrapKey(ciphertext, kek)
}

var (
	errCertificate = errors.DefineAborted("certificate", "invalid certificate `{id}`")
	errPrivateKey  = errors.DefineAborted("private_key", "invalid private key `{id}`")
)

func (k *keyVault) GetCertificate(ctx context.Context, id string) (cert *x509.Certificate, err error) {
	if v, err := k.certErrCache.Get(id); err == nil {
		return nil, v.(error)
	}
	if v, err := k.certCache.Get(id); err == nil {
		return v.(*x509.Certificate), nil
	}
	defer func() {
		if err != nil {
			k.certCache.Remove(id)
			k.certErrCache.Set(id, err)
		} else {
			k.certCache.Set(id, cert)
			k.certErrCache.Remove(id)
		}
	}()

	res, err := k.certs.GetCertificateWithContext(ctx, &acm.GetCertificateInput{
		CertificateArn: aws.String(id),
	})
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode([]byte(*res.Certificate))
	if certBlock == nil {
		return nil, errCertificate.WithAttributes("id", id)
	}
	return x509.ParseCertificate(certBlock.Bytes)
}

func (k *keyVault) ExportCertificate(ctx context.Context, id string) (cert *tls.Certificate, err error) {
	if v, err := k.certExportErrCache.Get(id); err == nil {
		return nil, v.(error)
	}
	if v, err := k.certExportCache.Get(id); err == nil {
		return v.(*tls.Certificate), nil
	}
	defer func() {
		if err != nil {
			k.certExportCache.Remove(id)
			k.certExportErrCache.Set(id, err)
		} else {
			k.certExportCache.Set(id, cert)
			k.certExportErrCache.Remove(id)
		}
	}()

	passphrase := []byte(random.String(16))
	res, err := k.certs.ExportCertificateWithContext(ctx, &acm.ExportCertificateInput{
		CertificateArn: aws.String(id),
		Passphrase:     passphrase,
	})
	if err != nil {
		return nil, err
	}

	cert = new(tls.Certificate)
	certBlock, _ := pem.Decode([]byte(*res.Certificate))
	if certBlock == nil {
		return nil, errCertificate.WithAttributes("id", id)
	}
	cert.Certificate = [][]byte{certBlock.Bytes}
	if res.CertificateChain != nil {
		chain := []byte(*res.CertificateChain)
		for {
			certBlock, rest := pem.Decode(chain)
			if certBlock == nil {
				break
			}
			cert.Certificate = append(cert.Certificate, certBlock.Bytes)
			chain = rest
		}
	}
	privateKeyBlock, _ := pem.Decode([]byte(*res.PrivateKey))
	if privateKeyBlock == nil {
		return nil, errPrivateKey.WithAttributes("id", id)
	}
	key, err := pkcs8.ParsePKCS8PrivateKey(privateKeyBlock.Bytes, passphrase)
	if err != nil {
		return nil, errPrivateKey.WithAttributes("id").WithCause(err)
	}
	cert.PrivateKey = stdcrypto.PrivateKey(key)
	return
}
