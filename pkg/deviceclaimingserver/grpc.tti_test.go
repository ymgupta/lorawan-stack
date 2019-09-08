// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	. "go.thethings.network/lorawan-stack/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/pkg/deviceclaimingserver/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestAuthorizeApplication(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	redisClient, redisFlush := test.NewRedis(t, "deviceclaimingserver_test", "applications", "authorized")
	defer redisFlush()
	defer redisClient.Close()
	authorizedApplicationsRegistry := &redis.AuthorizedApplicationRegistry{Redis: redisClient}

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	test.Must(New(c, &Config{
		AuthorizedApplications: authorizedApplicationsRegistry,
	}))
	test.Must(c.Start(), nil)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)

	client := ttnpb.NewEndDeviceClaimingServerClient(c.LoopbackConn())

	ids := ttnpb.ApplicationIdentifiers{
		ApplicationID: "foo",
	}

	// Application not authorized yet.
	_, err := client.UnauthorizeApplication(ctx, &ids)
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Authorize application.
	_, err = client.AuthorizeApplication(ctx, &ttnpb.AuthorizeApplicationRequest{
		ApplicationIdentifiers: ids,
		APIKey:                 "test",
	})
	a.So(err, should.BeNil)

	// Authorize application again with new API key.
	_, err = client.AuthorizeApplication(ctx, &ttnpb.AuthorizeApplicationRequest{
		ApplicationIdentifiers: ids,
		APIKey:                 "test-new",
	})
	a.So(err, should.BeNil)

	// Unauthorize application.
	_, err = client.UnauthorizeApplication(ctx, &ids)
	a.So(err, should.BeNil)

	// Application not authorized anymore.
	_, err = client.UnauthorizeApplication(ctx, &ids)
	a.So(errors.IsNotFound(err), should.BeTrue)
}

type sourceNSAddrKeyType struct{}
type sourceASAddrKeyType struct{}
type targetNSAddrKeyType struct{}
type targetASAddrKeyType struct{}

var (
	sourceNSAddrKey sourceNSAddrKeyType
	sourceASAddrKey sourceASAddrKeyType
	targetNSAddrKey targetNSAddrKeyType
	targetASAddrKey targetASAddrKeyType
)

var (
	errNotFound = errors.DefineNotFound("not_found", "not found")
	errRequest  = errors.DefineInvalidArgument("request", "invalid request")
)

func TestClaim(t *testing.T) {
	sourceTenantIDs := ttipb.TenantIdentifiers{
		TenantID: "source-tenant",
	}
	sourceAppIDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "source-app",
	}
	sourceDevIDs := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: sourceAppIDs,
		DeviceID:               "source-device",
		JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}
	targetTenantIDs := tenant.FromContext(test.Context())
	targetAppIDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "target-app",
	}
	targetDevIDs := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: targetAppIDs,
		DeviceID:               "target-device",
		JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}

	for _, tc := range []struct {
		Name              string
		Request           *ttnpb.ClaimEndDeviceRequest
		ApplicationRights map[string]*ttnpb.Rights

		GetIdentifiersForEndDeviceEUIsFunc func(context.Context, *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, ...grpc.CallOption) (*ttipb.TenantIdentifiers, error)
		GetEndDeviceIdentifiersForEUIsFunc func(context.Context, *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error)
		GetAuthorizedApplicationFunc       func(context.Context, ttnpb.ApplicationIdentifiers, []string) (*ttipb.ApplicationAPIKey, error)

		GetEndDeviceFunc   func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		JsGetEndDeviceFunc func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		NsGetEndDeviceFunc func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)
		AsGetEndDeviceFunc func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)

		DeleteEndDeviceFunc   func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
		JsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
		NsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)
		AsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)

		CreateEndDeviceFunc func(context.Context, *ttnpb.CreateEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		JsSetEndDeviceFunc  func(context.Context, *ttnpb.SetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		NsSetEndDeviceFunc  func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)
		AsSetEndDeviceFunc  func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)

		ErrorAssertion     func(t *testing.T, err error) bool
		ExpectSuccessEvent bool
		ExpectFailEvent    bool
	}{
		{
			Name: "InsufficientTargetRights",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "SourceTenantNotFound",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				return nil, errNotFound
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name: "SourceDeviceNotFound",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				return nil, errNotFound
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name: "ApplicationNotAuthorized",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(context.Context, ttnpb.ApplicationIdentifiers, []string) (*ttipb.ApplicationAPIKey, error) {
				return nil, errNotFound
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name: "InsufficientSourceRights",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "ClaimAuthenticationCode/None",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
				unique.ID(tenant.NewContext(context.Background(), sourceTenantIDs), sourceAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			GetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers:    sourceDevIDs,
					ClaimAuthenticationCode: nil,
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectFailEvent: true,
		},
		{
			Name: "ClaimAuthenticationCode/TooEarly",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
				unique.ID(tenant.NewContext(context.Background(), sourceTenantIDs), sourceAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			GetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				validFrom := time.Now().Add(1 * time.Hour)
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						ValidFrom: &validFrom,
					},
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectFailEvent: true,
		},
		{
			Name: "ClaimAuthenticationCode/TooLate",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
				unique.ID(tenant.NewContext(context.Background(), sourceTenantIDs), sourceAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			GetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				validTo := time.Now().Add(-1 * time.Hour)
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						ValidTo: &validTo,
					},
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectFailEvent: true,
		},
		{
			Name: "ClaimAuthenticationCode/Mismatch",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
				unique.ID(tenant.NewContext(context.Background(), sourceTenantIDs), sourceAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			GetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: []byte{0xff, 0xff},
					},
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectFailEvent: true,
		},
		{
			Name: "Success",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs: targetAppIDs,
				TargetDeviceID:       targetDevIDs.DeviceID,
			},
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(test.Context(), targetAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
				unique.ID(tenant.NewContext(context.Background(), sourceTenantIDs), sourceAppIDs): {Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				}},
			},
			GetIdentifiersForEndDeviceEUIsFunc: func(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound
				}
				return &ttipb.ApplicationAPIKey{
					ApplicationIDs: sourceAppIDs,
					APIKey:         "test",
				}, nil
			},
			GetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers:     sourceDevIDs,
					Name:                     "test",
					Description:              "test test",
					NetworkServerAddress:     ctx.Value(sourceNSAddrKey).(string),
					ApplicationServerAddress: ctx.Value(sourceASAddrKey).(string),
					JoinServerAddress:        "joinserver:1234",
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: []byte{0x01, 0x02},
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				}, nil
			},
			NsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					LoRaWANVersion:       ttnpb.MAC_V1_0_3,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_3_REV_A,
					FrequencyPlanID:      "EU_863_870",
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
					},
				}, nil
			},
			AsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
					},
				}, nil
			},
			DeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return ttnpb.Empty, nil
			},
			JsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return ttnpb.Empty, nil
			},
			NsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return ttnpb.Empty, nil
			},
			AsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound
				}
				return ttnpb.Empty, nil
			},
			CreateEndDeviceFunc: func(ctx context.Context, in *ttnpb.CreateEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers:     targetDevIDs,
						Name:                     "test",
						Description:              "test test",
						ApplicationServerAddress: ctx.Value(targetASAddrKey).(string),
						NetworkServerAddress:     ctx.Value(targetNSAddrKey).(string),
						JoinServerAddress:        "joinserver:1234",
					}) {
					return nil, errRequest
				}
				return &in.EndDevice, nil
			},
			JsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers:     targetDevIDs,
						ApplicationServerAddress: ctx.Value(targetASAddrKey).(string),
						NetworkServerAddress:     ctx.Value(targetNSAddrKey).(string),
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value: []byte{0x01, 0x02},
						},
						RootKeys: &ttnpb.RootKeys{
							RootKeyID: "test",
							AppKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					}) {
					return nil, errRequest
				}
				return &in.EndDevice, nil
			},
			NsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers: targetDevIDs,
						FrequencyPlanID:      "EU_863_870",
						LoRaWANVersion:       ttnpb.MAC_V1_0_3,
						LoRaWANPHYVersion:    ttnpb.PHY_V1_0_3_REV_A,
						MACSettings: &ttnpb.MACSettings{
							Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						},
					}) {
					return nil, errRequest
				}
				return &in.EndDevice, nil
			},
			AsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers: targetDevIDs,
						Formatters: &ttnpb.MessagePayloadFormatters{
							UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
							DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						},
					}) {
					return nil, errRequest
				}
				return &in.EndDevice, nil
			},
			ExpectSuccessEvent: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx, cancelCtx := context.WithCancel(log.NewContext(test.Context(), test.GetLogger(t)))
			defer cancelCtx()

			var successEvents, failEvents uint32
			defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{
				PublishFunc: func(ev events.Event) {
					switch name := ev.Name(); name {
					case "dcs.end_device.claim.success":
						atomic.AddUint32(&successEvents, 1)
					case "dcs.end_device.claim.fail":
						atomic.AddUint32(&failEvents, 1)
					}
				},
			})()
			defer func() {
				var expectedSuccessEvents, expectedFailEvents uint32
				if tc.ExpectSuccessEvent {
					expectedSuccessEvents = 1
				}
				if tc.ExpectFailEvent {
					expectedFailEvents = 1
				}
				a.So(atomic.LoadUint32(&successEvents), should.Equal, expectedSuccessEvents)
				a.So(atomic.LoadUint32(&failEvents), should.Equal, expectedFailEvents)
			}()

			sourceMockNS, sourceNSAddr := startMockNS(t, ctx)
			sourceMockNS.GetFunc = tc.NsGetEndDeviceFunc
			sourceMockNS.DeleteFunc = tc.NsDeleteEndDeviceFunc
			sourceMockAS, sourceASAddr := startMockAS(t, ctx)
			sourceMockAS.GetFunc = tc.AsGetEndDeviceFunc
			sourceMockAS.DeleteFunc = tc.AsDeleteEndDeviceFunc

			targetMockNS, targetNSAddr := startMockNS(t, ctx)
			targetMockNS.SetFunc = tc.NsSetEndDeviceFunc
			targetMockAS, targetASAddr := startMockAS(t, ctx)
			targetMockAS.SetFunc = tc.AsSetEndDeviceFunc

			dcs := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{
					ServiceBase: config.ServiceBase{
						GRPC: config.GRPC{
							AllowInsecureForCredentials: true,
						},
					},
				}),
				&Config{
					AuthorizedApplications: &mockAuthorizedApplicationsRegistry{
						GetFunc: tc.GetAuthorizedApplicationFunc,
					},
				},
				WithTenantRegistry(&mockTenantRegistry{
					GetIdentifiersForEndDeviceEUIsFunc: tc.GetIdentifiersForEndDeviceEUIsFunc,
				}),
				WithDeviceRegistry(&mockDeviceRegistry{
					CreateFunc:                tc.CreateEndDeviceFunc,
					GetFunc:                   tc.GetEndDeviceFunc,
					GetIdentifiersForEUIsFunc: tc.GetEndDeviceIdentifiersForEUIsFunc,
					DeleteFunc:                tc.DeleteEndDeviceFunc,
				}),
				WithJsDeviceRegistry(&mockJsDeviceRegistry{
					GetFunc:    tc.JsGetEndDeviceFunc,
					SetFunc:    tc.JsSetEndDeviceFunc,
					DeleteFunc: tc.JsDeleteEndDeviceFunc,
				}),
			)).(*DeviceClaimingServer)

			dcs.AddContextFiller(func(ctx context.Context) context.Context {
				ctx = rights.NewContext(ctx, rights.Rights{
					ApplicationRights: tc.ApplicationRights,
				})
				ctx = test.ContextWithT(ctx, t)
				ctx = context.WithValue(ctx, sourceNSAddrKey, sourceNSAddr)
				ctx = context.WithValue(ctx, sourceASAddrKey, sourceASAddr)
				ctx = context.WithValue(ctx, targetNSAddrKey, targetNSAddr)
				ctx = context.WithValue(ctx, targetASAddrKey, targetASAddr)
				return ctx
			})
			test.Must(dcs.Start(), nil)
			defer dcs.Close()

			mustHavePeer(ctx, dcs.Component, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)

			client := ttnpb.NewEndDeviceClaimingServerClient(dcs.LoopbackConn())
			callOpt := grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:  "Bearer",
				AuthValue: "test",
			})

			req := tc.Request
			req.TargetNetworkServerAddress = targetNSAddr
			req.TargetApplicationServerAddress = targetASAddr
			ids, err := client.Claim(ctx, req, callOpt)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				return
			}
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			if !a.So(*ids, should.Resemble, targetDevIDs) {
				t.FailNow()
			}
		})
	}
}
