// Copyright © 2020 The Things Industries B.V.

package eventserver_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/eventserver"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

var ConsumerIDs = [...]string{
	"stream",
}

func NewProtoEvent(ctx context.Context, ids ttnpb.Identifiers, data interface{}, rights ...ttnpb.Right) *ttnpb.Event {
	testName := test.MustTFromContext(ctx).Name()
	return test.Must(events.Proto(events.New(
		ctx, testName, testName,
		events.WithIdentifiers(ids), events.WithData(data), events.WithVisibility(rights...),
	))).(*ttnpb.Event)
}

func MakeEventData(t *testing.T) interface{} {
	return struct {
		Name string
	}{
		Name: t.Name(),
	}
}

var _ EventQueue = MockEventQueue{}

// MockEventQueue is a mock EventQueue used for testing.
type MockEventQueue struct {
	AddFunc func(context.Context, *ttnpb.Event) error
	PopFunc func(context.Context, string, func(context.Context, *ttnpb.Event) error) error
}

// Add calls AddFunc if set and panics otherwise.
func (m MockEventQueue) Add(ctx context.Context, ev *ttnpb.Event) error {
	if m.AddFunc == nil {
		panic("Add called, but not set")
	}
	return m.AddFunc(ctx, ev)
}

// Pop calls PopFunc if set and panics otherwise.
func (m MockEventQueue) Pop(ctx context.Context, id string, f func(context.Context, *ttnpb.Event) error) error {
	if m.PopFunc == nil {
		panic("Pop called, but not set")
	}
	return m.PopFunc(ctx, id, f)
}

type EventQueueAddRequest struct {
	Context  context.Context
	Event    *ttnpb.Event
	Response chan<- error
}

func MakeEventQueueAddChFunc(reqCh chan<- EventQueueAddRequest) func(context.Context, *ttnpb.Event) error {
	return func(ctx context.Context, ev *ttnpb.Event) error {
		respCh := make(chan error)
		reqCh <- EventQueueAddRequest{
			Context:  ctx,
			Event:    ev,
			Response: respCh,
		}
		return <-respCh
	}
}

type EventQueuePopRequest struct {
	Context  context.Context
	ID       string
	Func     func(context.Context, *ttnpb.Event) error
	Response chan<- error
}

func MakeEventQueuePopChFunc(reqCh chan<- EventQueuePopRequest) func(context.Context, string, func(context.Context, *ttnpb.Event) error) error {
	return func(ctx context.Context, id string, f func(context.Context, *ttnpb.Event) error) error {
		respCh := make(chan error)
		reqCh <- EventQueuePopRequest{
			Context:  ctx,
			ID:       id,
			Func:     f,
			Response: respCh,
		}
		return <-respCh
	}
}

type EventQueueEnvironment struct {
	Add <-chan EventQueueAddRequest
	Pop <-chan EventQueuePopRequest
}

func newMockEventQueue(t *testing.T) (EventQueue, EventQueueEnvironment, func()) {
	t.Helper()

	addCh := make(chan EventQueueAddRequest)
	popCh := make(chan EventQueuePopRequest)
	return &MockEventQueue{
			AddFunc: MakeEventQueueAddChFunc(addCh),
			PopFunc: MakeEventQueuePopChFunc(popCh),
		}, EventQueueEnvironment{
			Add: addCh,
			Pop: popCh,
		},
		func() {
			select {
			case <-addCh:
				t.Error("EventQueue.Add call missed")
			default:
				close(addCh)
			}
			select {
			case <-popCh:
				t.Error("EventQueue.Pop call missed")
			default:
				close(popCh)
			}
		}
}

type TestEnvironment struct {
	Cluster struct {
		Auth    <-chan test.ClusterAuthRequest
		GetPeer <-chan test.ClusterGetPeerRequest
	}
	SubscribeHandler events.Handler
	IngestQueue      *EventQueueEnvironment
}

func StartTest(t *testing.T, esConf Config, timeout time.Duration) (*EventServer, context.Context, TestEnvironment, func()) {
	t.Helper()

	authCh := make(chan test.ClusterAuthRequest)
	getPeerCh := make(chan test.ClusterGetPeerRequest)

	var env TestEnvironment
	env.Cluster.Auth = authCh
	env.Cluster.GetPeer = getPeerCh

	wg := &sync.WaitGroup{}
	var closeFuncs []func()
	if esConf.IngestQueue == nil {
		m, mEnv, closeM := newMockEventQueue(t)
		esConf.IngestQueue = m
		env.IngestQueue = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}
	if esConf.Subscriber == nil {
		wg.Add(1)
		esConf.Subscriber = &test.MockEventPubSub{
			SubscribeFunc: func() func(s string, h events.Handler) error {
				var mu sync.Mutex
				var wasCalled bool
				return func(s string, h events.Handler) error {
					defer wg.Done()

					mu.Lock()
					defer mu.Unlock()

					a := assertions.New(t)
					a.So(s, should.Equal, "**")
					if !a.So(wasCalled, should.BeFalse) {
						t.Error("Subscriber.Subscribe called more than once")
					}
					wasCalled = true
					return nil
				}
			}(),
		}
	}

	es := test.Must(New(
		componenttest.NewComponent(
			t,
			&component.Config{},
			component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
				return &test.MockCluster{
					AuthFunc:    test.MakeClusterAuthChFunc(authCh),
					GetPeerFunc: test.MakeClusterGetPeerChFunc(getPeerCh),
					JoinFunc:    test.ClusterJoinNilFunc,
					WithVerifiedSourceFunc: func(ctx context.Context) context.Context {
						return clusterauth.NewContext(ctx, nil)
					},
				}, nil
			}),
		),
		&esConf,
	)).(*EventServer)
	componenttest.StartComponent(t, es.Component)
	wg.Wait()

	ctx := test.ContextWithTB(test.Context(), t)
	ctx = log.NewContext(ctx, es.Logger())
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return es, ctx, env, func() {
		es.Close()
		cancel()
		for _, f := range closeFuncs {
			f()
		}
	}
}

func PerformConsumerTest(t *testing.T, id string, timeout time.Duration, f func(context.Context, *EventServer, TestEnvironment, func(context.Context, *ttnpb.Event) (<-chan error, bool)) bool) {
	t.Helper()
	a := assertions.New(t)

	es, ctx, env, stop := StartTest(t, DefaultConfig, timeout)
	t.Cleanup(stop)

	popCh := env.IngestQueue.Pop
	env.IngestQueue = nil // Remove reference, such that caller could not directly interact with it.
	seen := map[string]struct{}{}
	outerT := t
	if !a.So(f(ctx, es, env, func(ctx context.Context, ev *ttnpb.Event) (<-chan error, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)

		for {
			var req EventQueuePopRequest
			if !a.So(test.WaitContext(ctx, func() { req = <-popCh }), should.BeTrue) {
				t.Error("Timed out while waiting for IngestQueue.Pop to be called")
				return nil, false
			}
			if !test.AllTrue(
				a.So(req.ID, should.BeIn, ConsumerIDs),
				a.So(seen, should.NotContainKey, req.ID),
			) {
				return nil, false
			}
			t.Logf("IngestQueue.Pop called by consumer with ID %s", req.ID)
			if req.ID != id {
				outerT.Cleanup(func() { req.Response <- nil })
				seen[req.ID] = struct{}{}
				continue
			}
			errCh := make(chan error, 1)
			go func() {
				err := req.Func(ctx, ev)
				errCh <- err
				close(errCh)
				req.Response <- err
			}()
			return errCh, true
		}
	}), should.BeTrue) {
		t.Error("Consumer test handler failed")
		return
	}
	es.Close()
	for len(seen) < len(ConsumerIDs) {
		var req EventQueuePopRequest
		if !a.So(test.WaitContext(ctx, func() { req = <-popCh }), should.BeTrue) {
			t.Error("Timed out while waiting for IngestQueue.Pop to be called")
			return
		}
		if !test.AllTrue(
			a.So(req.ID, should.BeIn, ConsumerIDs),
			a.So(seen, should.NotContainKey, req.ID),
		) {
			return
		}
		t.Logf("IngestQueue.Pop called by consumer with ID %s", req.ID)
		t.Cleanup(func() { req.Response <- nil })
		seen[req.ID] = struct{}{}
		continue
	}
}

func AssertErrorIsNil(t *testing.T, err error) bool {
	t.Helper()
	return assertions.New(t).So(err, should.BeNil)
}

func AssertErrorIsInvalidArgument(t *testing.T, err error) bool {
	t.Helper()
	return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
}

func AssertErrorIsUnauthenticated(t *testing.T, err error) bool {
	t.Helper()
	return assertions.New(t).So(errors.IsUnauthenticated(err), should.BeTrue)
}

func AssertReceiveError(ctx context.Context, ch <-chan error, assert func(*testing.T, error) bool) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	a := assertions.New(t)
	var err error
	if !a.So(test.WaitContext(ctx, func() { err = <-ch }), should.BeTrue) {
		return false
	}
	return a.So(assert(t, err), should.BeTrue)
}

func AssertIngest(ctx context.Context, ingest func(context.Context, *ttnpb.Event) (<-chan error, bool), ev *ttnpb.Event, handle func() bool, assertError func(*testing.T, error) bool) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	a := assertions.New(t)

	errCh, ok := ingest(ctx, ev)
	if !a.So(ok, should.BeTrue) {
		t.Error("Consumer failed to ingest event")
		return false
	}
	if !a.So(handle(), should.BeTrue) {
		return false
	}
	return a.So(AssertReceiveError(ctx, errCh, assertError), should.BeTrue)
}

func NewISPeer(ctx context.Context, is interface {
	ttnpb.ApplicationAccessServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, is, ttnpb.RegisterApplicationAccessServer)).(cluster.Peer)
}

func AssertListRights(ctx context.Context, env TestEnvironment, assert func(ctx context.Context, ids ttnpb.Identifiers) bool, rights ...ttnpb.Right) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	a := assertions.New(t)

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer close(listRightsCh)

	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			return test.AllTrue(
				a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS),
				a.So(ids, should.BeNil),
			)
		},
		test.ClusterGetPeerResponse{
			Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
				ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
			}),
		},
	), should.BeTrue) {
		return false
	}
	return a.So(test.AssertListRightsRequest(ctx, listRightsCh, assert, rights...), should.BeTrue)
}
