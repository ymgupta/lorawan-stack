// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/kms"
	"go.thethings.network/lorawan-stack/pkg/log"
)

type awsPart struct {
	PartNumber           string `dynamodbav:"part_number"`
	PSKRootKeysEncrypted []byte `dynamodbav:"psk_root_keys_encrypted"`
}

// NewAWSKeyLoader returns a key loader that loads keys from Amazon Web Services.
func NewAWSKeyLoader(region, table string) KeyLoaderFunc {
	return func(ctx context.Context) (map[string]Key, error) {
		ctx = log.NewContextWithField(ctx, "namespace", "cryptoserver/provisioners/microchip/aws")
		logger := log.FromContext(ctx).WithField("parts_table", table)
		awsconfig := aws.NewConfig()
		if region != "" {
			awsconfig = awsconfig.WithRegion(region)
		}
		ses, err := session.NewSession(awsconfig)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize session")
			return nil, err
		}
		scanInput := &dynamodb.ScanInput{
			TableName: aws.String(table),
		}
		var parts []awsPart
		err = dynamodb.New(ses).ScanPagesWithContext(ctx, scanInput, func(page *dynamodb.ScanOutput, lastPage bool) bool {
			for _, item := range page.Items {
				var part awsPart
				if err := dynamodbattribute.UnmarshalMap(item, &part); err != nil {
					logger.WithError(err).Warn("Failed to unmarshal item")
					continue
				}
				parts = append(parts, part)
			}
			return true
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to scan table")
			return nil, err
		}
		keys := make(map[string]Key)
		kmsClient := kms.New(ses)
		for _, part := range parts {
			logger := logger.WithField("part_number", part.PartNumber)
			response, err := kmsClient.DecryptWithContext(ctx, &kms.DecryptInput{
				CiphertextBlob: part.PSKRootKeysEncrypted,
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to decrypt key")
				continue
			}
			var key Key
			copy(key[:], response.Plaintext)
			keys[part.PartNumber] = key
		}
		return keys, nil
	}
}
