// Copyright © 2019 The Things Industries B.V.

package awsmetrics

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/marketplacemetering"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

// Reporter is a license.MeteringReporter that reports the stats
// to the AWS Marketplace metering service.
type Reporter struct {
	config  *ttipb.MeteringConfiguration_AWS
	service *marketplacemetering.MarketplaceMetering
}

var (
	reportBackoff = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

	errMeteringServiceUnavailable = errors.DefineUnavailable("metering_service_unavailable", "AWS Marketplace metering service unavailable")
)

// Report implements license.MeteringReporter.
func (r *Reporter) Report(ctx context.Context, data *ttipb.MeteringData) (err error) {
retryAttempt:
	for _, backoff := range reportBackoff {
		request := &marketplacemetering.MeterUsageInput{
			ProductCode: aws.String(r.config.GetSKU()),
			Timestamp:   aws.Time(time.Now()),
		}
		request.UsageDimension, request.UsageQuantity = ComputeUsage(data)
		_, err = r.service.MeterUsageWithContext(ctx, request)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case marketplacemetering.ErrCodeInternalServiceErrorException:
				time.Sleep(random.Jitter(backoff, 0.1))
				continue retryAttempt
			}
		}
		return err
	}
	return errMeteringServiceUnavailable.WithCause(err)
}

// New returns a new license.MeteringReporter that reports the metrics to the AWS Marketplace Metering Service.
func New(config *ttipb.MeteringConfiguration_AWS) (*Reporter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	service := marketplacemetering.New(sess)
	return &Reporter{
		config:  config,
		service: service,
	}, nil
}

// ComputeUsage computes usage.
func ComputeUsage(data *ttipb.MeteringData) (*string, *int64) {
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
		return aws.String("Up20000devices"), aws.Int64((endDeviceCount + 99) / 100)
	}
}
