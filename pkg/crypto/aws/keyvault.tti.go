// Copyright © 2019 The Things Industries B.V.

package aws

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

const (
	kekTTL        = (1 << 4) * time.Minute
	kekErrTTL     = (1 << 3) * time.Minute
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

var (
	errSecretContent  = errors.DefineAborted("secret_content", "invalid secret content in `{id}`")
	errSecretNotFound = errors.DefineNotFound("secret_not_found", "secret `{id}` not found")
)

func (k *keyVault) loadKEK(ctx context.Context, kekLabel string) (kek []byte, err error) {
	id := kekLabel
	if k.secretIDPrefix != "" {
		id = fmt.Sprintf("%s/%s", k.secretIDPrefix, id)
	}
	if v, err := k.kekErrCache.Get(id); err == nil {
		crypto.RegisterCacheHit(ctx, "aws_kek")
		return nil, v.(error)
	}
	if v, err := k.kekCache.Get(id); err == nil {
		crypto.RegisterCacheHit(ctx, "aws_kek")
		return v.([]byte), nil
	}
	defer func() {
		crypto.RegisterCacheMiss(ctx, "aws_kek")
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
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				return nil, errSecretNotFound.WithCause(err).WithAttributes("id", id)
			}
		}
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

var errCertificate = errors.DefineAborted("certificate", "invalid certificate `{id}`")

type certificateSecret struct {
	Certificate string `json:"certificate"`
	Key         string `json:"key,omitempty"`
}

func (k *keyVault) GetCertificate(ctx context.Context, id string) (cert *x509.Certificate, err error) {
	if k.secretIDPrefix != "" {
		id = fmt.Sprintf("%s/%s", k.secretIDPrefix, id)
	}
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
	res, err := k.secrets.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	if res.SecretString == nil {
		return nil, errSecretContent.WithAttributes("id", id)
	}
	var secret certificateSecret
	if err := json.Unmarshal([]byte(*res.SecretString), &secret); err != nil {
		return nil, errSecretContent.WithCause(err).WithAttributes("id", id)
	}
	certBlock, _ := pem.Decode([]byte(secret.Certificate))
	if certBlock == nil {
		return nil, errCertificate.WithAttributes("id", id)
	}
	return x509.ParseCertificate(certBlock.Bytes)
}

func (k *keyVault) ExportCertificate(ctx context.Context, id string) (cert *tls.Certificate, err error) {
	if k.secretIDPrefix != "" {
		id = fmt.Sprintf("%s/%s", k.secretIDPrefix, id)
	}
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
	res, err := k.secrets.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	if res.SecretString == nil {
		return nil, errSecretContent.WithAttributes("id", id)
	}
	var secret certificateSecret
	if err := json.Unmarshal([]byte(*res.SecretString), &secret); err != nil {
		return nil, errSecretContent.WithCause(err).WithAttributes("id", id)
	}
	pair, err := tls.X509KeyPair([]byte(secret.Certificate), []byte(secret.Key))
	if err != nil {
		return nil, err
	}
	cert = &pair
	return
}
