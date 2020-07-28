// Copyright © 2020 The Things Industries B.V.

package eventserver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/eventserver"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestStream(t *testing.T) {
	PerformConsumerTest(t, DefaultConfig.Consumers.StreamGroup, (1<<9)*test.Delay, func(ctx context.Context, es *EventServer, env TestEnvironment, ingest func(context.Context, *ttnpb.Event) (<-chan error, bool)) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		a.So(AssertIngest(ctx, ingest, &ttnpb.Event{}, func() bool { return true }, AssertErrorIsNil), should.BeTrue)

		appID := &ttnpb.EntityIdentifiers{
			Ids: &ttnpb.EntityIdentifiers_ApplicationIDs{
				ApplicationIDs: &ttnpb.ApplicationIdentifiers{
					ApplicationID: "test-app",
				},
			},
		}
		devID := &ttnpb.EntityIdentifiers{
			Ids: &ttnpb.EntityIdentifiers_DeviceIDs{
				DeviceIDs: &ttnpb.EndDeviceIdentifiers{
					DeviceID: "test-dev",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "test-app-dev",
					},
				},
			},
		}

		const authType = "Bearer"
		const authValue = "auth-token"

		assertListRights := func(ctx context.Context, expectedIDs ttnpb.Identifiers, rights ...ttnpb.Right) bool {
			t := test.MustTFromContext(ctx)
			t.Helper()
			return assertions.New(t).So(AssertListRights(ctx, env, func(ctx context.Context, ids ttnpb.Identifiers) bool {
				md := rpcmetadata.FromIncomingContext(ctx)
				return test.AllTrue(
					a.So(md.AuthType, should.Equal, authType),
					a.So(md.AuthValue, should.Equal, authValue),
					a.So(ids, should.Resemble, expectedIDs),
				)
			}, rights...), should.BeTrue)
		}

		var appIDStreams []ttnpb.Events_StreamClient
		var devIDStreams []ttnpb.Events_StreamClient
		timePtr := func(v time.Time) *time.Time { return &v }
		for _, tc := range []struct {
			Request        *ttnpb.StreamEventsRequest
			Rights         *rights.Rights
			Handler        func(ctx context.Context, cl ttnpb.Events_StreamClient) bool
			ErrorAssertion func(t *testing.T, err error) bool
		}{
			{
				Request:        &ttnpb.StreamEventsRequest{},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Tail: 42,
				},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					After: timePtr(time.Now()),
				},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Tail:  42,
					After: timePtr(time.Now()),
				},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						{
							Ids: &ttnpb.EntityIdentifiers_ApplicationIDs{},
						},
					},
				},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						{
							Ids: &ttnpb.EntityIdentifiers_ClientIDs{
								ClientIDs: &ttnpb.ClientIdentifiers{},
							},
						},
					},
				},
				ErrorAssertion: AssertErrorIsInvalidArgument,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						appID,
					},
				},
				Handler: func(ctx context.Context, cl ttnpb.Events_StreamClient) bool {
					a := assertions.New(test.MustTFromContext(ctx))
					return a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
						func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
							return test.AllTrue(
								a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS),
								a.So(ids, should.BeNil),
							)
						},
						test.ClusterGetPeerResponse{
							Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{}),
						},
					), should.BeTrue)
				},
				ErrorAssertion: AssertErrorIsUnauthenticated,
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						appID,
					},
				},
				Rights: &rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, appID): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
						},
					},
				},
				Handler: func(ctx context.Context, cl ttnpb.Events_StreamClient) bool {
					ev, err := cl.Recv()
					if a.So(err, should.BeNil) {
						a.So(ev.Name, should.Equal, "events.stream.start")
					}
					appIDStreams = append(appIDStreams, cl)
					return true
				},
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						devID,
					},
				},
				Rights: &rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, devID.GetDeviceIDs().ApplicationIdentifiers): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
						},
					},
				},
				Handler: func(ctx context.Context, cl ttnpb.Events_StreamClient) bool {
					ev, err := cl.Recv()
					if a.So(err, should.BeNil) {
						a.So(ev.Name, should.Equal, "events.stream.start")
					}
					devIDStreams = append(devIDStreams, cl)
					return true
				},
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						appID,
						devID,
					},
				},
				Rights: &rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, appID): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
						},
						unique.ID(ctx, devID.GetDeviceIDs().ApplicationIdentifiers): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
						},
					},
				},
				Handler: func(ctx context.Context, cl ttnpb.Events_StreamClient) bool {
					ev, err := cl.Recv()
					if a.So(err, should.BeNil) {
						a.So(ev.Name, should.Equal, "events.stream.start")
					}
					appIDStreams = append(appIDStreams, cl)
					devIDStreams = append(devIDStreams, cl)
					return true
				},
			},
			{
				Request: &ttnpb.StreamEventsRequest{
					Identifiers: []*ttnpb.EntityIdentifiers{
						devID.GetDeviceIDs().ApplicationIdentifiers.EntityIdentifiers(),
					},
				},
				Rights: &rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, devID.GetDeviceIDs().ApplicationIdentifiers): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
						},
					},
				},
				Handler: func(ctx context.Context, cl ttnpb.Events_StreamClient) bool {
					ev, err := cl.Recv()
					if a.So(err, should.BeNil) {
						a.So(ev.Name, should.Equal, "events.stream.start")
					}
					devIDStreams = append(devIDStreams, cl)
					return true
				},
			},
		} {
			t.Run(fmt.Sprintf("open stream/tail:%v,after:%v,identifiers:%v,rights:%+v", tc.Request.Tail, tc.Request.After, tc.Request.Identifiers, tc.Rights), func(t *testing.T) {
				ctx := test.ContextWithTB(ctx, t)
				a := assertions.New(t)

				var opts []grpc.CallOption
				if tc.Rights != nil {
					opts = append(opts, grpc.PerRPCCredentials(rpcmetadata.MD{
						AuthType:      authType,
						AuthValue:     authValue,
						AllowInsecure: true,
					}))
				}

				cl, err := ttnpb.NewEventsClient(es.LoopbackConn()).Stream(ctx, tc.Request, opts...)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to open stream: %s", err)
				}

				if tc.Rights != nil {
					for _, reqID := range tc.Request.Identifiers {
						var rights map[string]*ttnpb.Rights
						switch ids := reqID.Ids.(type) {
						case *ttnpb.EntityIdentifiers_ApplicationIDs:
							rights = tc.Rights.ApplicationRights
						case *ttnpb.EntityIdentifiers_ClientIDs:
							rights = tc.Rights.ClientRights
						case *ttnpb.EntityIdentifiers_DeviceIDs:
							reqID = ids.DeviceIDs.ApplicationIdentifiers.EntityIdentifiers()
							rights = tc.Rights.ApplicationRights
						case *ttnpb.EntityIdentifiers_GatewayIDs:
							rights = tc.Rights.GatewayRights
						case *ttnpb.EntityIdentifiers_OrganizationIDs:
							rights = tc.Rights.OrganizationRights
						case *ttnpb.EntityIdentifiers_UserIDs:
							rights = tc.Rights.UserRights
						default:
							t.Fatalf("Unmatched identifier type %T", reqID.Ids)
						}
						if !a.So(assertListRights(ctx, reqID.Identifiers(), rights[unique.ID(ctx, reqID.Identifiers())].GetRights()...), should.BeTrue) {
							t.Fatalf("ListRights assertion failed")
						}
					}
				}
				if tc.Handler != nil {
					a.So(tc.Handler(ctx, cl), should.BeTrue)
				}
				if tc.ErrorAssertion != nil {
					ev, err := cl.Recv()
					if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
						a.So(ev, should.BeNil)
					}
				}
			})
		}

		for _, tc := range []struct {
			Name           string
			Identifiers    ttnpb.Identifiers
			RequiredRights []ttnpb.Right
			ErrorAssertion func(t *testing.T, err error) bool
			Streams        []ttnpb.Events_StreamClient
		}{
			{
				Identifiers: appID,
				RequiredRights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
				},
				ErrorAssertion: AssertErrorIsNil,
				Streams:        appIDStreams,
			},
			{
				Identifiers: appID,
				RequiredRights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
				},
				ErrorAssertion: AssertErrorIsNil,
				Streams:        appIDStreams,
			},
			{
				Identifiers:    appID,
				ErrorAssertion: AssertErrorIsNil,
				Streams:        appIDStreams,
			},
			{
				Identifiers:    devID,
				ErrorAssertion: AssertErrorIsNil,
				Streams:        devIDStreams,
			},
			{
				Identifiers: devID,
				RequiredRights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
				},
				ErrorAssertion: AssertErrorIsNil,
				Streams:        devIDStreams,
			},
		} {
			t.Run(fmt.Sprintf("ingest event/identifiers:%v,required rights:%v", tc.Identifiers.Identifiers(), tc.RequiredRights), func(t *testing.T) {
				type rightTestCase struct {
					Rights  []ttnpb.Right
					Visible bool
				}
				rightTCs := []rightTestCase{
					{
						Visible: len(tc.RequiredRights) == 0,
					},
				}
				if len(tc.RequiredRights) > 0 {
					rightTCs = append(rightTCs,
						rightTestCase{
							Rights:  tc.RequiredRights,
							Visible: true,
						},
						rightTestCase{
							Rights:  []ttnpb.Right{ttnpb.RIGHT_SEND_INVITES},
							Visible: false,
						},
						rightTestCase{
							Rights:  []ttnpb.Right{tc.RequiredRights[0], ttnpb.RIGHT_SEND_INVITES},
							Visible: true,
						},
					)
				}
				if len(tc.RequiredRights) > 1 {
					rightTCs = append(rightTCs, rightTestCase{
						Rights:  tc.RequiredRights[0:1],
						Visible: true,
					})
				}
				for _, rightTC := range rightTCs {
					t.Run(fmt.Sprintf("stream rights:%v", rightTC.Rights), func(t *testing.T) {
						ctx := test.ContextWithTB(ctx, t)
						a := assertions.New(t)

						expectedEv := NewProtoEvent(ctx, tc.Identifiers, MakeEventData(t), tc.RequiredRights...)
						a.So(AssertIngest(ctx, ingest, deepcopy.Copy(expectedEv).(*ttnpb.Event), func() bool {
							var recvs []func() (*ttnpb.Event, error)
							for _, cl := range tc.Streams {
								if len(tc.RequiredRights) > 0 {
									reqIDs := tc.Identifiers.Identifiers()
									switch ids := reqIDs.(type) {
									case *ttnpb.ApplicationIdentifiers:
									case *ttnpb.ClientIdentifiers:
									case *ttnpb.EndDeviceIdentifiers:
										reqIDs = &ids.ApplicationIdentifiers
									case *ttnpb.GatewayIdentifiers:
									case *ttnpb.OrganizationIdentifiers:
									case *ttnpb.UserIdentifiers:
									default:
										t.Fatalf("Unmatched identifier type %T", reqIDs)
									}
									if !a.So(assertListRights(ctx, reqIDs, rightTC.Rights...), should.BeTrue) {
										t.Fatalf("ListRights assertion failed")
									}
								}
								if !rightTC.Visible {
									continue
								}
								recvs = append(recvs, cl.Recv)
							}
							for _, recv := range recvs {
								ev, err := recv()
								if !a.So(err, should.BeNil) {
									return false
								}
								a.So(ev, should.Resemble, expectedEv)
							}
							return true
						}, tc.ErrorAssertion), should.BeTrue)
					})
				}
			})
		}
		cancel()
		for _, cl := range append(appIDStreams, devIDStreams...) {
			ev, err := cl.Recv()
			if !a.So(errors.IsCanceled(err), should.BeTrue) {
				return false
			}
			a.So(ev, should.BeNil)
		}
		return true
	})
}
