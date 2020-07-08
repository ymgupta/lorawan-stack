// Copyright © 2020 The Things Industries B.V.

package eventserver

import (
	"go.thethings.network/lorawan-stack/v3/pkg/events"
)

// ConsumerConfig represents the consumer configuration.
type ConsumerConfig struct {
	StreamGroup string
}

// GroupNames returns names of all configured non-empty consumer groups.
func (conf ConsumerConfig) GroupNames() []string {
	if conf.StreamGroup == "" {
		return nil
	}
	return []string{
		conf.StreamGroup,
	}
}

// Config represents EventServer configuration.
type Config struct {
	Subscriber  events.Subscriber `name:"-"`
	IngestQueue EventQueue        `name:"-"`
	Consumers   ConsumerConfig    `name:"-"`
}

// DefDefaultConfig is the default Event Server config.
var DefaultConfig = Config{
	Consumers: ConsumerConfig{
		StreamGroup: "stream",
	},
}
