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
		ses.Config.MergeIn(&aws.Config{
			Credentials: stscreds.NewCredentials(
				ses, settings.AWSIoT.AssumeRole.ARN,
				func(p *stscreds.AssumeRoleProvider) {
					if d := settings.AWSIoT.AssumeRole.SessionDuration; d != nil {
						p.Duration = *d
					}
					if id := settings.AWSIoT.AssumeRole.ExternalID; id != "" {
						p.ExternalID = aws.String(id)
					}
				},
			),
		})
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
	topics := provider.Topics(target)
	if defaultIntegration := settings.AWSIoT.GetDefault(); defaultIntegration != nil {
		topics = &defaultIntegrationTopics{
			baseTopic: fmt.Sprintf(defaultIntegrationBaseTopicFormat, defaultIntegration.StackName),
		}
	}
	return mqtt.OpenConnection(ctx, mqttSettings, topics)
}

const defaultIntegrationBaseTopicFormat = "thethings/lorawan/%s"

type defaultIntegrationTopics struct {
	baseTopic string
}

func (t defaultIntegrationTopics) GetBaseTopic() string {
	return t.baseTopic
}

func (t defaultIntegrationTopics) GetUplinkMessage() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "uplink",
	}
}

func (t defaultIntegrationTopics) GetJoinAccept() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "join",
	}
}

func (t defaultIntegrationTopics) GetDownlinkAck() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/ack",
	}
}

func (t defaultIntegrationTopics) GetDownlinkNack() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/nack",
	}
}

func (t defaultIntegrationTopics) GetDownlinkSent() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/sent",
	}
}

func (t defaultIntegrationTopics) GetDownlinkFailed() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/failed",
	}
}

func (t defaultIntegrationTopics) GetDownlinkQueued() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/queued",
	}
}

func (t defaultIntegrationTopics) GetLocationSolved() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "location/solved",
	}
}

func (t defaultIntegrationTopics) GetServiceData() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "service/data",
	}
}

func (t defaultIntegrationTopics) GetDownlinkPush() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/push",
	}
}

func (t defaultIntegrationTopics) GetDownlinkReplace() *ttnpb.ApplicationPubSub_Message {
	return &ttnpb.ApplicationPubSub_Message{
		Topic: "downlink/replace",
	}
}

func init() {
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_AWSIoT{}, impl{})
}
