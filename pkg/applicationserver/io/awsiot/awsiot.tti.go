// Copyright © 2019 The Things Industries B.V.

package awsiot

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
	asaws "go.thethings.network/lorawan-stack/pkg/applicationserver/aws"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// NewSubscription returns a new subscription that publishes to AWS IoT.
func NewSubscription(ctx context.Context, config asaws.Config) (*io.Subscription, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/awsiot")

	awsconfig := aws.NewConfig()
	if config.Region != "" {
		awsconfig = awsconfig.WithRegion(config.Region)
	}
	ses, err := session.NewSession(awsconfig)
	if err != nil {
		return nil, err
	}
	res, err := iot.New(ses).DescribeEndpointWithContext(ctx, &iot.DescribeEndpointInput{})
	if err != nil {
		return nil, err
	}

	client := iotdataplane.New(ses, awsconfig.WithEndpoint(*res.EndpointAddress))

	sub := io.NewSubscription(ctx, "awsiot", nil)
	go func() {
		logger := log.FromContext(ctx)
		for {
			select {
			case <-ctx.Done():
				return

			case msg := <-sub.Up():
				logger := logger.WithField("device_uid", unique.ID(msg.Context, msg.EndDeviceIdentifiers))
				data, err := jsonpb.TTN().Marshal(msg.ApplicationUp)
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal message")
					continue
				}
				var upType string
				switch msg.Up.(type) {
				case *ttnpb.ApplicationUp_UplinkMessage:
					upType = "up"
				case *ttnpb.ApplicationUp_JoinAccept:
					upType = "join"
				case *ttnpb.ApplicationUp_DownlinkAck:
					upType = "down/ack"
				case *ttnpb.ApplicationUp_DownlinkNack:
					upType = "down/nack"
				case *ttnpb.ApplicationUp_DownlinkSent:
					upType = "down/sent"
				case *ttnpb.ApplicationUp_DownlinkFailed:
					upType = "down/failed"
				case *ttnpb.ApplicationUp_DownlinkQueued:
					upType = "down/queued"
				case *ttnpb.ApplicationUp_LocationSolved:
					upType = "location"
				default:
					continue
				}
				topic := fmt.Sprintf("lorawan/%s/%s/things/%s/%s",
					tenant.FromContext(msg.Context).TenantID,
					msg.ApplicationID,
					msg.DeviceID,
					upType,
				)
				logger = logger.WithField("topic", topic)
				_, err = client.PublishWithContext(ctx, &iotdataplane.PublishInput{
					Payload: data,
					Topic:   aws.String(topic),
				})
				if err != nil {
					logger.WithError(err).Warn("Failed to publish message")
					continue
				}
				logger.Debug("Published message")
			}
		}
	}()
	return sub, nil
}
