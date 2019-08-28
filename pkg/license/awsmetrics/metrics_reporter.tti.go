// Copyright © 2019 The Things Industries B.V.

package awsmetrics

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/marketplacemetering"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

type reporter struct {
	config  *ttipb.MeteringConfiguration_AWS
	service *marketplacemetering.MarketplaceMetering
}

// Report implements license.MetricsReporter.
func (r *reporter) Report(ctx context.Context, data *ttipb.MeteringData) error {
	request := &marketplacemetering.MeterUsageInput{
		ProductCode: aws.String(r.config.GetSKU()),
		Timestamp:   aws.Time(time.Now()),
	}
	request.UsageDimension, request.UsageQuantity = computeUsage(data)
	response, err := r.service.MeterUsageWithContext(ctx, request)
	logger := log.FromContext(ctx)
	if err != nil {
		logger.WithError(err).Warn("Failed to transmit the metering record")
	} else {
		logger.WithField("metering_record_id", response.MeteringRecordId).Debug("Metering record sent")
	}
	return err
}

// New returns a new license.MetricsReporter that reports the metrics to the AWS Marketplace Metering Service.
func New(config *ttipb.MeteringConfiguration_AWS) (license.MetricsReporter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	service := marketplacemetering.New(sess)
	return &reporter{
		config:  config,
		service: service,
	}, nil
}

func computeUsage(data *ttipb.MeteringData) (*string, *int64) {
	var endDeviceCount int64
	for _, tenant := range data.GetTenants() {
		endDeviceCount += int64(tenant.GetTotals().GetEndDevices())
	}
	switch {
	case endDeviceCount <= 1000:
		return aws.String("1000devices"), aws.Int64(1)
	case endDeviceCount <= 2000:
		return aws.String("2000devices"), aws.Int64(1)
	case endDeviceCount <= 4000:
		return aws.String("4000devices"), aws.Int64(1)
	case endDeviceCount <= 6500:
		return aws.String("6500devices"), aws.Int64(1)
	case endDeviceCount <= 10000:
		return aws.String("10000devices"), aws.Int64(1)
	case endDeviceCount <= 15000:
		return aws.String("15000devices"), aws.Int64(1)
	case endDeviceCount <= 20000:
		return aws.String("20000devices"), aws.Int64(1)
	default:
		return aws.String("Over20000devices"), aws.Int64(endDeviceCount / 100)
	}
}
