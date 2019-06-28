// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/datastore"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"go.thethings.network/lorawan-stack/pkg/log"
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
		logger := log.FromContext(ctx).WithField("gcp_project_id", projectID)
		partsClient, err := datastore.NewClient(ctx, projectID)
		if err != nil {
			return nil, err
		}
		var parts []gcpPart
		if _, err := partsClient.GetAll(ctx, datastore.NewQuery(kind), &parts); err != nil {
			return nil, err
		}
		keys := make(map[string]Key)
		for _, part := range parts {
			logger.WithField("part_number", part.Name.Name).Debug("Load part")
			storageClient, err := storage.NewClient(ctx)
			if err != nil {
				return nil, err
			}
			reader, err := storageClient.Bucket(part.KeyBucket).Object(part.KeyDataObject).NewReader(ctx)
			if err != nil {
				return nil, err
			}
			encryptedKey, err := ioutil.ReadAll(reader)
			if err != nil {
				return nil, err
			}
			kmsClient, err := kms.NewKeyManagementClient(ctx)
			if err != nil {
				return nil, err
			}
			response, err := kmsClient.AsymmetricDecrypt(ctx, &kmspb.AsymmetricDecryptRequest{
				Name:       part.KEK,
				Ciphertext: encryptedKey,
			})
			if err != nil {
				return nil, err
			}
			var key Key
			copy(key[:], response.Plaintext)
			keys[part.Name.Name] = key
		}
		return keys, nil
	}
}
