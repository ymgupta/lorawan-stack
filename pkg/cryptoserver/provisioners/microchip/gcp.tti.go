// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/datastore"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type gcpPart struct {
	Name          *datastore.Key `datastore:"__key__"`
	KeyBucket     string         `datastore:"key_bucket,noindex"`
	KeyDataObject string         `datastore:"key_data_object,noindex"`
	KEK           string         `datastore:"kek,noindex"`
}

// NewGCPKeyLoader returns a key loader that loads keys from Google Cloud Platform.
// Parts are stored as Datastore entities with the given kind. The part number is the key of the entity.
// The encrypted parent key is loaded from the bucket `key_bucket` and data object `key_data_object`.
// The parent key is decrypted using the asymmetric KMS key `kek`.
func NewGCPKeyLoader(projectID, kind string) KeyLoaderFunc {
	return func(ctx context.Context) (map[string]Key, error) {
		ctx = log.NewContextWithField(ctx, "namespace", "cryptoserver/provisioners/microchip/gcp")
		logger := log.FromContext(ctx).WithField("project_id", projectID)
		partsClient, err := datastore.NewClient(ctx, projectID)
		if err != nil {
			logger.WithError(err).Warn("Failed to create Datastore client")
			return nil, err
		}
		defer partsClient.Close()
		var parts []gcpPart
		if _, err := partsClient.GetAll(ctx, datastore.NewQuery(kind), &parts); err != nil {
			logger.WithError(err).WithField("kind", kind).Warn("Failed to get parts")
			return nil, err
		}
		keys := make(map[string]Key)
		for _, part := range parts {
			logger := logger.WithField("part_number", part.Name.Name)
			logger.Debug("Load part")
			storageClient, err := storage.NewClient(ctx)
			if err != nil {
				logger.WithError(err).Warn("Failed to create Storage client")
				return nil, err
			}
			defer storageClient.Close()
			reader, err := storageClient.Bucket(part.KeyBucket).Object(part.KeyDataObject).NewReader(ctx)
			if err != nil {
				logger.WithError(err).WithFields(log.Fields(
					"key_bucket", part.KeyBucket,
					"key_data_object", part.KeyDataObject,
				)).Warn("Failed to create data object reader")
				return nil, err
			}
			defer reader.Close()
			encryptedKey, err := ioutil.ReadAll(reader)
			if err != nil {
				logger.WithError(err).Warn("Failed to read data object")
				return nil, err
			}
			kmsClient, err := kms.NewKeyManagementClient(ctx)
			if err != nil {
				logger.WithError(err).Warn("Failed to create KMS client")
				return nil, err
			}
			defer kmsClient.Close()
			response, err := kmsClient.AsymmetricDecrypt(ctx, &kmspb.AsymmetricDecryptRequest{
				Name:       part.KEK,
				Ciphertext: encryptedKey,
			})
			if err != nil {
				logger.WithError(err).WithField("kek", part.KEK).Warn("Failed to decrypt KEK")
				return nil, err
			}
			var key Key
			copy(key[:], response.Plaintext)
			keys[part.Name.Name] = key
		}
		return keys, nil
	}
}
