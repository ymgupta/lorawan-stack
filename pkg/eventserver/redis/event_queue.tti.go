// Copyright © 2020 The Things Industries B.V.

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EventQueue is an implementation of eventserver.EventQueue.
type EventQueue struct {
	Redis  *ttnredis.Client
	MaxLen int64
	ID     string
}

const (
	queueKey   = "queue"
	payloadKey = "payload"
)

// Init initializes the event queue.
// It must be called at least once before using the queue.
func (q *EventQueue) Init(groups ...string) error {
	for _, group := range groups {
		if err := q.Redis.XGroupCreateMkStream(q.Redis.Key(queueKey), group, "$").Err(); err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
			return ttnredis.ConvertError(err)
		}
	}
	return nil
}

// Add adds event to queue.
func (q *EventQueue) Add(ctx context.Context, pb *ttnpb.Event) error {
	s, err := ttnredis.MarshalProto(pb)
	if err != nil {
		return err
	}
	return ttnredis.ConvertError(q.Redis.XAdd(&redis.XAddArgs{
		Stream:       q.Redis.Key(queueKey),
		MaxLenApprox: q.MaxLen,
		Values: map[string]interface{}{
			payloadKey: s,
		},
	}).Err())
}

var (
	errInvalidKeyValueType = errors.DefineInvalidArgument("value_type", "invalid value type for key `{key}`")
	errNoPayload           = errors.DefineInvalidArgument("no_payload", "no payload")
)

// Pop pops an event from the queue, which has not yet been delivered to any consumer within group identified by id.
// If no such event is available, Pop blocks until either there is one or ctx.Deadline() is reached.
// Pop does not respect ctx.Done() directly.
func (q *EventQueue) Pop(ctx context.Context, id string, f func(context.Context, *ttnpb.Event) error) error {
	var timeout time.Duration
	dl, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(dl)
	}
	rets, err := q.Redis.XReadGroup(&redis.XReadGroupArgs{
		Group:    id,
		Consumer: q.ID,
		Streams:  []string{q.Redis.Key(queueKey), ">"},
		Count:    1,
		Block:    timeout,
	}).Result()
	if err != nil && err != redis.Nil {
		return ttnredis.ConvertError(err)
	}
	for _, ret := range rets {
		for _, msg := range ret.Messages {
			v, ok := msg.Values[payloadKey]
			if !ok {
				return errNoPayload
			}
			s, ok := v.(string)
			if !ok {
				return errInvalidKeyValueType.WithAttributes("key", payloadKey)
			}

			pb := &ttnpb.Event{}
			if err := ttnredis.UnmarshalProto(s, pb); err != nil {
				return err
			}
			if err := f(ctx, pb); err != nil {
				return err
			}
			_, err = q.Redis.XAck(ret.Stream, id, msg.ID).Result()
			return ttnredis.ConvertError(err)
		}
	}
	return nil
}
