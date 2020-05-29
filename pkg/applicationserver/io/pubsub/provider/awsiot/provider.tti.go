// Copyright © 2020 The Things Industries B.V.

// Package awsiot implements an AWS IoT pub/sub provider.
package awsiot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	sigv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/iot"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider/mqtt"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type impl struct {
}

// OpenConnection implements provider.Provider using the MQTT driver.
func (impl) OpenConnection(ctx context.Context, target provider.Target) (pc *provider.Connection, err error) {
	settings, ok := target.GetProvider().(*ttnpb.ApplicationPubSub_AWSIoT)
	if !ok {
		panic("wrong provider type provided to OpenConnection")
	}
	awsConfig := aws.NewConfig()
	if settings.AWSIoT.Region != "" {
		awsConfig = awsConfig.WithRegion(settings.AWSIoT.Region)
	}
	if settings.AWSIoT.AccessKey != nil {
		awsConfig = awsConfig.WithCredentials(credentials.NewStaticCredentials(
			settings.AWSIoT.AccessKey.AccessKeyID,
			settings.AWSIoT.AccessKey.SecretAccessKey,
			settings.AWSIoT.AccessKey.SessionToken,
		))
	}
	ses, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}
	if settings.AWSIoT.AssumeRole != nil {
		awsConfig = awsConfig.WithCredentials(stscreds.NewCredentials(
			ses, settings.AWSIoT.AssumeRole.ARN,
			func(p *stscreds.AssumeRoleProvider) {
				if settings.AWSIoT.AssumeRole.SessionDuration != nil {
					p.Duration = *settings.AWSIoT.AssumeRole.SessionDuration
				}
				if settings.AWSIoT.AssumeRole.ExternalID != "" {
					p.ExternalID = aws.String(settings.AWSIoT.AssumeRole.ExternalID)
				}
			},
		))
		ses.Config.MergeIn(awsConfig)
	}
	endpointAddress := settings.AWSIoT.EndpointAddress
	if endpointAddress == "" {
		res, err := iot.New(ses).DescribeEndpointWithContext(ctx, &iot.DescribeEndpointInput{})
		if err != nil {
			return nil, err
		}
		endpointAddress = *res.EndpointAddress
	}
	brokerAddress := fmt.Sprintf("wss://%s/mqtt", endpointAddress)
	req, _ := http.NewRequest("GET", brokerAddress, nil)
	signer := sigv4.NewSigner(ses.Config.Credentials, func(s *sigv4.Signer) {
		s.DisableHeaderHoisting = true
	})
	_, err = signer.Sign(req, nil, "iotdevicegateway", settings.AWSIoT.Region, time.Now())
	if err != nil {
		return nil, err
	}
	headers := make(map[string]string)
	for _, key := range []string{"Authorization", "X-Amz-Date", "X-Amz-Security-Token"} {
		if value := req.Header.Get(key); value != "" {
			headers[key] = value
		}
	}
	mqttSettings := &ttnpb.ApplicationPubSub_MQTT{
		MQTT: &ttnpb.ApplicationPubSub_MQTTProvider{
			ServerURL: brokerAddress,
			Headers:   headers,
		},
	}
	return mqtt.OpenConnection(ctx, mqttSettings, target)
}

func init() {
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_AWSIoT{}, impl{})
}
