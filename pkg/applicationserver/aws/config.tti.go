// Copyright © 2019 The Things Industries B.V.

package aws

// Config represents configuration for AWS IoT.
type Config struct {
	Region string `name:"region" description:"AWS region (optional)"`
	IoT    struct {
		Telemetry bool `name:"telemetry" description:"Enable publishing telemetry to AWS IoT"`
	} `name:"iot" description:"AWS IoT configuration"`
}
