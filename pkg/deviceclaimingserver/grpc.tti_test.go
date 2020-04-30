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
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	. "go.thethings.network/lorawan-stack/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/pkg/deviceclaimingserver/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/discover"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/web/oauthclient"
	"google.golang.org/grpc"
)

func TestAuthorizeApplication(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	redisClient, redisFlush := test.NewRedis(t, "deviceclaimingserver_test", "applications", "authorized")
	defer redisFlush()
	defer redisClient.Close()
	authorizedApplicationsRegistry := &redis.AuthorizedApplicationRegistry{Redis: redisClient}

	c := componenttest.NewComponent(t, &component.Config{})
	test.Must(New(c, &Config{
		OAuth: oauthclient.Config{
			AuthorizeURL: "http://localhost/oauth/authorize",
			TokenURL:     "http://localhost/token",
			ClientID:     "test",
			ClientSecret: "test",
		},
		AuthorizedApplications: authorizedApplicationsRegistry,
	}))

	componenttest.StartComponent(t, c)
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
type createsKeyType struct{}
type deletesKeyType struct{}

var (
	sourceNSAddrKey sourceNSAddrKeyType
	sourceASAddrKey sourceASAddrKeyType
	targetNSAddrKey targetNSAddrKeyType
	targetASAddrKey targetASAddrKeyType
	createsKey      createsKeyType
	deletesKey      deletesKeyType
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

		CreateEndDeviceFunc   func(context.Context, *ttnpb.CreateEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		GetEndDeviceFunc      func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		DeleteEndDeviceFunc   func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
		JsGetEndDeviceFunc    func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		JsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)

		SourceNsGetEndDeviceFunc    func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)
		SourceAsGetEndDeviceFunc    func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)
		SourceNsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)
		SourceAsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)

		TargetJsSetEndDeviceFunc    func(context.Context, *ttnpb.SetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
		TargetNsSetEndDeviceFunc    func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)
		TargetAsSetEndDeviceFunc    func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)
		TargetNsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)
		TargetAsDeleteEndDeviceFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)

		ExpectCreates       int64
		ExpectSourceDeletes int64
		ExpectSuccessEvent  bool
		ExpectAbortEvent    bool
		ExpectFailEvent     bool

		ErrorAssertion         func(t *testing.T, err error) bool
		FailEventDataAssertion func(t *testing.T, dev *ttnpb.EndDevice) bool
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
				return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(context.Context, ttnpb.ApplicationIdentifiers, []string) (*ttipb.ApplicationAPIKey, error) {
				return nil, errNotFound.New()
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers:    sourceDevIDs,
					ClaimAuthenticationCode: nil,
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectAbortEvent: true,
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			ExpectAbortEvent: true,
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			ExpectAbortEvent: true,
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
				}, nil
			},
			JsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "FFFF",
					},
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsAborted(err), should.BeTrue)
			},
			ExpectAbortEvent: true,
		},
		{
			Name: "TargetASUnavailable",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs:           targetAppIDs,
				TargetDeviceID:                 targetDevIDs.DeviceID,
				TargetApplicationServerAddress: "invalid:42424",
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "0102",
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				}, nil
			},
			SourceNsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			SourceAsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
					},
				}, nil
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.NotBeNil)
			},
			ExpectAbortEvent: true,
		},
		{
			Name: "SuccessWithASDeleteFail",
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "0102",
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				}, nil
			},
			SourceNsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			SourceAsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			JsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			SourceNsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			SourceAsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				panic("something went wrong")
			},
			CreateEndDeviceFunc: func(ctx context.Context, in *ttnpb.CreateEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetJsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers:     targetDevIDs,
						ApplicationServerAddress: ctx.Value(targetASAddrKey).(string),
						NetworkServerAddress:     ctx.Value(targetNSAddrKey).(string),
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value: "0102",
						},
						RootKeys: &ttnpb.RootKeys{
							RootKeyID: "test",
							AppKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					}) {
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetNsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetAsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers: targetDevIDs,
						Formatters: &ttnpb.MessagePayloadFormatters{
							UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
							DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						},
					}) {
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			ExpectCreates:       4,
			ExpectSourceDeletes: 3,
			ExpectSuccessEvent:  true,
		},
		{
			Name: "RollbackCreates",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs:         targetAppIDs,
				TargetDeviceID:               targetDevIDs.DeviceID,
				InvalidateAuthenticationCode: true,
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "0102",
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				}, nil
			},
			SourceNsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			SourceAsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				switch tenant.FromContext(ctx) {
				case sourceTenantIDs:
					if !a.So(*in, should.Resemble, sourceDevIDs) {
						return nil, errNotFound.New()
					}
				case targetTenantIDs:
					if !a.So(*in, should.Resemble, targetDevIDs) {
						return nil, errNotFound.New()
					}
				default:
					t.Fatal("Unknown tenant")
				}
				return ttnpb.Empty, nil
			},
			JsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				switch tenant.FromContext(ctx) {
				case sourceTenantIDs:
					if !a.So(*in, should.Resemble, sourceDevIDs) {
						return nil, errNotFound.New()
					}
				case targetTenantIDs:
					if !a.So(*in, should.Resemble, targetDevIDs) {
						return nil, errNotFound.New()
					}
				default:
					t.Fatal("Unknown tenant")
				}
				return ttnpb.Empty, nil
			},
			SourceNsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			SourceAsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			CreateEndDeviceFunc: func(ctx context.Context, in *ttnpb.CreateEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetJsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers:     targetDevIDs,
						ApplicationServerAddress: ctx.Value(targetASAddrKey).(string),
						NetworkServerAddress:     ctx.Value(targetNSAddrKey).(string),
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value:   "0102",
							ValidTo: in.EndDevice.ClaimAuthenticationCode.ValidTo,
						},
						RootKeys: &ttnpb.RootKeys{
							RootKeyID: "test",
							AppKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					}) {
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetNsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetAsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				panic("something went wrong")
			},
			TargetNsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(*in, should.Resemble, targetDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			TargetAsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(*in, should.Resemble, targetDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			ExpectCreates:       4,
			ExpectSourceDeletes: 4 + 3, // 4 source plus 3 rollback creates on AS create fail.
			ExpectAbortEvent:    true,
			ExpectFailEvent:     true,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.NotBeNil)
			},
			FailEventDataAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers:     sourceDevIDs,
					Name:                     "test",
					Description:              "test test",
					NetworkServerAddress:     dev.NetworkServerAddress,
					ApplicationServerAddress: dev.ApplicationServerAddress,
					JoinServerAddress:        "joinserver:1234",
					LoRaWANVersion:           ttnpb.MAC_V1_0_3,
					LoRaWANPHYVersion:        ttnpb.PHY_V1_0_3_REV_A,
					FrequencyPlanID:          "EU_863_870",
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
					},
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
					},
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "0102",
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				})
			},
		},
		{
			Name: "Success",
			Request: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIDs:         targetAppIDs,
				TargetDeviceID:               targetDevIDs.DeviceID,
				InvalidateAuthenticationCode: true,
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
					return nil, errNotFound.New()
				}
				return &sourceTenantIDs, nil
			},
			GetEndDeviceIdentifiersForEUIsFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.JoinEUI, should.Resemble, types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) ||
					!a.So(in.DevEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
					return nil, errNotFound.New()
				}
				return &sourceDevIDs, nil
			},
			GetAuthorizedApplicationFunc: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(ids, should.Resemble, sourceAppIDs) ||
					!a.So(paths, should.Resemble, []string{"api_key"}) {
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
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
					return nil, errNotFound.New()
				}
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: sourceDevIDs,
					ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
						Value: "0102",
					},
					RootKeys: &ttnpb.RootKeys{
						RootKeyID: "test",
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				}, nil
			},
			SourceNsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
			SourceAsGetEndDeviceFunc: func(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(in.EndDeviceIdentifiers, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
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
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			JsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			SourceNsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			SourceAsDeleteEndDeviceFunc: func(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
				test.MustIncrementContextCounter(ctx, deletesKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, sourceTenantIDs) ||
					!a.So(*in, should.Resemble, sourceDevIDs) {
					return nil, errNotFound.New()
				}
				return ttnpb.Empty, nil
			},
			CreateEndDeviceFunc: func(ctx context.Context, in *ttnpb.CreateEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetJsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers:     targetDevIDs,
						ApplicationServerAddress: ctx.Value(targetASAddrKey).(string),
						NetworkServerAddress:     ctx.Value(targetNSAddrKey).(string),
						ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
							Value:   "0102",
							ValidTo: in.EndDevice.ClaimAuthenticationCode.ValidTo,
						},
						RootKeys: &ttnpb.RootKeys{
							RootKeyID: "test",
							AppKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					}) {
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetNsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
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
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			TargetAsSetEndDeviceFunc: func(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
				test.MustIncrementContextCounter(ctx, createsKey, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(tenant.FromContext(ctx), should.Resemble, targetTenantIDs) ||
					!a.So(in.EndDevice, should.Resemble, ttnpb.EndDevice{
						EndDeviceIdentifiers: targetDevIDs,
						Formatters: &ttnpb.MessagePayloadFormatters{
							UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
							DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
						},
					}) {
					return nil, errRequest.New()
				}
				return &in.EndDevice, nil
			},
			ExpectCreates:       4,
			ExpectSourceDeletes: 4,
			ExpectSuccessEvent:  true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx, cancelCtx := context.WithCancel(log.NewContext(test.Context(), test.GetLogger(t)))
			defer cancelCtx()

			var successEvents, abortEvents, failEvents uint32
			defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{
				PublishFunc: func(ev events.Event) {
					switch name := ev.Name(); name {
					case "dcs.end_device.claim.success":
						atomic.AddUint32(&successEvents, 1)
					case "dcs.end_device.claim.abort":
						atomic.AddUint32(&abortEvents, 1)
					case "dcs.end_device.claim.fail":
						atomic.AddUint32(&failEvents, 1)
						if tc.FailEventDataAssertion != nil {
							a.So(tc.FailEventDataAssertion(t, ev.Data().(*ttnpb.EndDevice)), should.BeTrue)
						}
					}
				},
			})()
			defer func() {
				var expectedSuccessEvents, expectedAbortEvents, expectedFailEvents uint32
				if tc.ExpectSuccessEvent {
					expectedSuccessEvents = 1
				}
				if tc.ExpectAbortEvent {
					expectedAbortEvents = 1
				}
				if tc.ExpectFailEvent {
					expectedFailEvents = 1
				}
				a.So(atomic.LoadUint32(&successEvents), should.Equal, expectedSuccessEvents)
				a.So(atomic.LoadUint32(&abortEvents), should.Equal, expectedAbortEvents)
				a.So(atomic.LoadUint32(&failEvents), should.Equal, expectedFailEvents)
			}()

			var creates, deletes int64
			withTestCounters := func(ctx context.Context) context.Context {
				ctx = test.ContextWithT(ctx, t)
				ctx = test.ContextWithCounterRef(ctx, createsKey, &creates)
				ctx = test.ContextWithCounterRef(ctx, deletesKey, &deletes)
				return ctx
			}

			sourceMockNS, sourceNSAddr := startMockNS(ctx, rpcserver.WithContextFiller(withTestCounters))
			sourceMockNS.GetFunc = tc.SourceNsGetEndDeviceFunc
			sourceMockNS.DeleteFunc = tc.SourceNsDeleteEndDeviceFunc
			sourceMockAS, sourceASAddr := startMockAS(ctx, rpcserver.WithContextFiller(withTestCounters))
			sourceMockAS.GetFunc = tc.SourceAsGetEndDeviceFunc
			sourceMockAS.DeleteFunc = tc.SourceAsDeleteEndDeviceFunc

			targetMockNS, targetNSAddr := startMockNS(ctx, rpcserver.WithContextFiller(withTestCounters))
			targetMockNS.SetFunc = tc.TargetNsSetEndDeviceFunc
			targetMockNS.DeleteFunc = tc.TargetNsDeleteEndDeviceFunc
			targetMockAS, targetASAddr := startMockAS(ctx, rpcserver.WithContextFiller(withTestCounters))
			targetMockAS.SetFunc = tc.TargetAsSetEndDeviceFunc
			targetMockAS.DeleteFunc = tc.TargetAsDeleteEndDeviceFunc

			c := componenttest.NewComponent(t, &component.Config{
				ServiceBase: config.ServiceBase{
					GRPC: config.GRPC{
						AllowInsecureForCredentials: true,
					},
				},
			})
			dcs := test.Must(New(c,
				&Config{
					OAuth: oauthclient.Config{
						AuthorizeURL: "http://localhost/oauth/authorize",
						TokenURL:     "http://localhost/token",
						ClientID:     "test",
						ClientSecret: "test",
					},
					AuthorizedApplications: &mockAuthorizedApplicationsRegistry{
						GetFunc: tc.GetAuthorizedApplicationFunc,
					},
				},
				WithTenantRegistry(&mockTenantRegistry{
					GetIdentifiersForEndDeviceEUIsFunc: tc.GetIdentifiersForEndDeviceEUIsFunc,
				}),
				WithApplicationAccess(&mockApplicationAccess{
					ListRightsFunc: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, opts ...grpc.CallOption) (*ttnpb.Rights, error) {
						return tc.ApplicationRights[unique.ID(ctx, ids)], nil
					},
				}),
				WithDeviceRegistry(&mockDeviceRegistry{
					CreateFunc:                tc.CreateEndDeviceFunc,
					GetFunc:                   tc.GetEndDeviceFunc,
					GetIdentifiersForEUIsFunc: tc.GetEndDeviceIdentifiersForEUIsFunc,
					DeleteFunc:                tc.DeleteEndDeviceFunc,
				}),
				WithJsDeviceRegistry(&mockJsDeviceRegistry{
					GetFunc:    tc.JsGetEndDeviceFunc,
					SetFunc:    tc.TargetJsSetEndDeviceFunc,
					DeleteFunc: tc.JsDeleteEndDeviceFunc,
				}),
			)).(*DeviceClaimingServer)

			dcs.AddContextFiller(func(ctx context.Context) context.Context {
				ctx = test.ContextWithT(ctx, t)
				ctx = discover.WithTLSFallback(ctx, false)
				ctx = rights.NewContext(ctx, rights.Rights{
					ApplicationRights: tc.ApplicationRights,
				})
				ctx = context.WithValue(ctx, sourceNSAddrKey, sourceNSAddr)
				ctx = context.WithValue(ctx, sourceASAddrKey, sourceASAddr)
				ctx = context.WithValue(ctx, targetNSAddrKey, targetNSAddr)
				ctx = context.WithValue(ctx, targetASAddrKey, targetASAddr)
				return ctx
			})
			dcs.AddContextFiller(withTestCounters)
			componenttest.StartComponent(t, c)
			defer c.Close()

			mustHavePeer(ctx, dcs.Component, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)

			client := ttnpb.NewEndDeviceClaimingServerClient(dcs.LoopbackConn())
			callOpt := grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:  "Bearer",
				AuthValue: "test",
			})

			req := tc.Request
			if req.TargetNetworkServerAddress == "" {
				req.TargetNetworkServerAddress = targetNSAddr
			}
			if req.TargetApplicationServerAddress == "" {
				req.TargetApplicationServerAddress = targetASAddr
			}
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
			if !a.So(creates, should.Equal, tc.ExpectCreates) {
				t.Fatal("Unexpected number of creates")
			}
			if !a.So(deletes, should.Equal, tc.ExpectSourceDeletes) {
				t.Fatal("Unexpected number of deletes")
			}
		})
	}
}
