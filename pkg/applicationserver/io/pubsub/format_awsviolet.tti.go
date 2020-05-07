// Copyright © 2020 The Things Industries B.V.

package pubsub

import "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/formatters"

func init() {
	formats["awsviolet"] = Format{
		Formatter: formatters.AWSViolet,
		Name:      "AWS Violet",
	}
}
