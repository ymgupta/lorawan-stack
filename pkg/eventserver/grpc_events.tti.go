// Copyright © 2020 The Things Industries B.V.

package eventserver

import (
	"os"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights/rightsutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/metadata"
)

var errNoIdentifiers = errors.DefineInvalidArgument("no_identifiers", "no identifiers")

// Stream is called by subscribers.
func (es *EventServer) Stream(req *ttnpb.StreamEventsRequest, stream ttnpb.Events_StreamServer) error {
	if len(req.Identifiers) == 0 {
		return errNoIdentifiers
	}

	ctx := stream.Context()
	logger := log.FromContext(ctx)

	if err := rights.RequireAny(ctx, req.Identifiers...); err != nil {
		logger.WithError(err).Info("Request contains no rights for selected identifiers, drop stream")
		return err
	}

	ch := make(events.Channel, 8)
	handler := events.ContextHandler(ctx, ch)
	es.filter.Subscribe(ctx, req, handler)
	defer es.filter.Unsubscribe(ctx, req, handler)

	if req.Tail > 0 || req.After != nil {
		logger = logger.WithField("tail", req.Tail)
		if req.After != nil {
			logger = logger.WithField("after", *req.After)
		}
		warning.Add(ctx, "Historical events not implemented")
	}

	if err := stream.SendHeader(metadata.MD{}); err != nil {
		logger.WithError(err).Warn("Failed to send header, drop stream")
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if err := stream.Send(&ttnpb.Event{
		Name:           "events.stream.start",
		Time:           time.Now().UTC(),
		Identifiers:    req.Identifiers,
		Origin:         hostname,
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
	}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev := <-ch:
			logger := logger.WithField("event_name", ev.Name())

			isVisible, err := rightsutil.EventIsVisible(ctx, ev)
			if err != nil {
				logger.WithError(err).Error("Failed to determine visibility of event, skip")
				return err
			}
			if !isVisible {
				logger.Debug("Insufficient rights for event, skip")
				continue
			}
			pb, ok := ev.(protoEvent)
			if !ok {
				logger.Error("Invalid event type, skip")
				continue
			}
			logger.Debug("Send event on stream")
			if err := stream.Send(pb.proto); err != nil {
				return err
			}
		}
	}
}
