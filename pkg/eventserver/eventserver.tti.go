// Copyright © 2020 The Things Industries B.V.

// Package eventserver provides an Event Server implementation, which stores and processes events.
package eventserver

import (
	"context"
	"fmt"
	"time"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// EventQueue represents a fan-out event queue.
type EventQueue interface {
	// Add adds *ttnpb.Event to the queue.
	// Implementations must ensure that Add returns fast.
	Add(context.Context, *ttnpb.Event) error
	// Pop calls a function for each available *ttnpb.Event in the queue in order.
	// Implementations must respect ctx.Deadline() value on best-effort basis, if such is present.
	Pop(context.Context, string, func(context.Context, *ttnpb.Event) error) error
}

type taskRegistrar interface {
	RegisterTask(context.Context, string, component.TaskFunc, component.TaskRestart, ...time.Duration)
}

func registerEventQueueProcessor(ctx context.Context, r taskRegistrar, q EventQueue, id string, f func(context.Context, *ttnpb.Event) error, restart component.TaskRestart, backoff ...time.Duration) {
	r.RegisterTask(ctx, fmt.Sprintf("event_queue_processor:%s", id), func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if err := q.Pop(ctx, id, func(ctx context.Context, pb *ttnpb.Event) error {
				return f(log.NewContextWithField(ctx, "event_name", pb.Name), pb)
			}); err != nil {
				return err
			}
		}
	}, restart, backoff...)
}

// EventServer represents an event server.
type EventServer struct {
	*component.Component
	filter events.IdentifierFilter
}

// protoEvent represents an events.Event, which has an associated proto form.
type protoEvent struct {
	events.Event
	proto *ttnpb.Event
}

var errInvalidConfiguration = errors.DefineInvalidArgument("configuration", "invalid configuration")

// New returns new EventServer.
func New(c *component.Component, conf *Config) (*EventServer, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "eventserver")
	if err := license.RequireComponent(ctx, ttnpb.ClusterRole_EVENT_SERVER); err != nil {
		return nil, err
	}
	switch {
	case conf.IngestQueue == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("IngestQueue is not specified")))
	case conf.Subscriber == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("Subscriber is not specified")))
	case conf.Consumers.StreamGroup == "":
		panic(errInvalidConfiguration.WithCause(errors.New("Consumers.StreamGroup is not specified")))
	}

	filter := events.NewIdentifierFilter()
	c.RegisterTask(ctx, "ingest", func(ctx context.Context) error {
		logger := log.FromContext(ctx)
		return conf.Subscriber.Subscribe("**", events.HandlerFunc(func(ev events.Event) {
			pb, err := events.Proto(ev)
			if err != nil {
				logger.WithError(err).Error("Failed to encode event")
				return
			}
			if err := conf.IngestQueue.Add(ctx, pb); err != nil {
				logger.WithError(err).Error("Failed to enqueue event")
			}
		}))
	}, component.TaskRestartOnFailure)
	registerEventQueueProcessor(ctx, c, conf.IngestQueue, conf.Consumers.StreamGroup, func(ctx context.Context, pb *ttnpb.Event) error {
		ev, err := events.FromProto(pb)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to decode event")
			return nil
		}
		filter.Notify(protoEvent{
			Event: ev,
			proto: pb,
		})
		return nil
	}, component.TaskRestartOnFailure)

	es := &EventServer{
		Component: c,
		filter:    filter,
	}
	c.RegisterGRPC(es)
	return es, nil
}

// Roles implements rpcserver.Registerer.
func (es *EventServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_EVENT_SERVER}
}

// RegisterServices implements rpcserver.Registerer.
func (es *EventServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEventsServer(s, es)
}

// RegisterHandlers implements rpcserver.Registerer.
func (es *EventServer) RegisterHandlers(s *grpc_runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEventsHandler(es.Context(), s, conn)
}
