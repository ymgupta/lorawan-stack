// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcode"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/discover"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type endDeviceClaimingServer struct {
	DCS *DeviceClaimingServer
}

var (
	errParseQRCode             = errors.Define("parse_qr_code", "parse QR code failed")
	errQRCodeData              = errors.DefineInvalidArgument("qr_code_data", "invalid QR code data")
	errAuthorizationNotFound   = errors.DefineNotFound("application_not_authorized", "application not authorized")
	errPermissionDenied        = errors.DefinePermissionDenied("permission_denied", "permission denied")
	errClaimAuthenticationCode = errors.DefineAborted("claim_authentication_code", "invalid claim authentication code")
)

var (
	transferISPaths = [...]string{
		"application_server_address",
		"attributes",
		"description",
		"join_server_address",
		"locations",
		"name",
		"network_server_address",
		"service_profile_id",
		"version_ids",
	}
	transferJSPaths = [...]string{
		"application_server_address",
		"application_server_id",
		"application_server_kek_label",
		"claim_authentication_code",
		"last_dev_nonce",
		"last_join_nonce",
		"net_id",
		"network_server_address",
		"network_server_kek_label",
		"provisioner_id",
		"provisioning_data",
		"resets_join_nonces",
		"root_keys",
		"used_dev_nonces",
	}
	transferNSPaths = [...]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"max_frequency",
		"min_frequency",
		"multicast",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
		"version_ids",
	}
	transferASPaths = [...]string{
		"formatters",
		"version_ids",
	}
)

func getHost(address string) string {
	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err == nil {
			address = url.Host
		}
	}
	if strings.Contains(address, ":") {
		host, _, err := net.SplitHostPort(address)
		if err == nil {
			return host
		}
	}
	return address
}

// customSameHost allows for overriding determining whether the given addresses are on the same host.
// This is useful for testing.
var customSameHost func(address1, address2 string) bool

func sameHost(address1, address2 string) bool {
	if customSameHost != nil {
		return customSameHost(address1, address2)
	}
	return strings.EqualFold(getHost(address1), getHost(address2))
}

var dialTimeout = 5 * time.Second

func (s *endDeviceClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	targetCtx := ctx
	if err := rights.RequireApplication(targetCtx, req.TargetApplicationIDs,
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
	); err != nil {
		return nil, err
	}
	targetForwardAuth, err := rpcmetadata.WithForwardedAuth(targetCtx, s.DCS.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx)

	// Get the JoinEUI, DevEUI and claim authentication code from the request.
	var joinEUI, devEUI types.EUI64
	var authCode string
	switch source := req.SourceDevice.(type) {
	case *ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_:
		authIDs := source.AuthenticatedIdentifiers
		joinEUI, devEUI, authCode = authIDs.JoinEUI, authIDs.DevEUI, authIDs.AuthenticationCode
	case *ttnpb.ClaimEndDeviceRequest_QRCode:
		data, err := qrcode.Parse(source.QRCode)
		if err != nil {
			return nil, errParseQRCode.WithCause(err)
		}
		authIDs, ok := data.(qrcode.AuthenticatedEndDeviceIdentifiers)
		if !ok {
			return nil, errQRCodeData.New()
		}
		joinEUI, devEUI, authCode = authIDs.AuthenticatedEndDeviceIdentifiers()
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", req.SourceDevice))
	}
	logger = logger.WithFields(log.Fields(
		"source_join_eui", joinEUI,
		"source_dev_eui", devEUI,
	))

	// Get the source end device identifiers belonging to the JoinEUI and DevEUI.
	logger.Debug("Get source tenant identifiers by EUIs")
	sourceCtx := rights.NewContextWithCache(ctx)
	tenantRegistry, err := s.DCS.getTenantRegistry(sourceCtx)
	if err != nil {
		logger.WithError(err).Warn("Failed to get tenant registry")
		return nil, err
	}
	tenantIDs, err := tenantRegistry.GetIdentifiersForEndDeviceEUIs(sourceCtx, &ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest{
		JoinEUI: joinEUI,
		DevEUI:  devEUI,
	}, s.DCS.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to get source tenant identifiers by EUIs")
		return nil, err
	}
	logger.Debug("Get source end device identifiers by EUIs")
	sourceCtx = tenant.NewContext(sourceCtx, *tenantIDs)
	sourceIDs := &ttnpb.EndDeviceIdentifiers{
		JoinEUI: &joinEUI,
		DevEUI:  &devEUI,
	}
	sourceERClient, err := s.DCS.getDeviceRegistry(sourceCtx, sourceIDs)
	if err != nil {
		logger.WithError(err).Warn("Failed to get device registry")
		return nil, err
	}
	sourceIDs, err = sourceERClient.GetIdentifiersForEUIs(sourceCtx, &ttnpb.GetEndDeviceIdentifiersForEUIsRequest{
		JoinEUI: joinEUI,
		DevEUI:  devEUI,
	}, s.DCS.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to get source end device identifiers by EUIs")
		return nil, err
	}
	logger = logger.WithField("source_device_uid", unique.ID(sourceCtx, sourceIDs))

	// Load the API key of the authorized source application.
	logger.Debug("Load authorized source application API key and validate rights")
	app, err := s.DCS.authorizedAppsRegistry.Get(sourceCtx, sourceIDs.ApplicationIdentifiers, []string{"api_key"})
	if err != nil {
		logger.WithError(err).Warn("Failed to load authorized source application API key")
		if errors.IsNotFound(err) {
			return nil, errPermissionDenied.WithCause(err)
		}
		return nil, err
	}
	sourceMD := rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     app.APIKey,
		AllowInsecure: s.DCS.AllowInsecureForCredentials(),
	}
	sourceCallOpts := []grpc.CallOption{
		grpc.PerRPCCredentials(sourceMD),
	}
	sourceDialOpts := append([]grpc.DialOption(nil), rpcclient.DefaultDialOptions(sourceCtx)...)
	sourceDialOpts = append(sourceDialOpts, grpc.WithBlock(), grpc.FailOnNonTempDialError(true))
	sourceTLSConfig, err := s.DCS.GetTLSClientConfig(sourceCtx)
	if err != nil {
		return nil, err
	}

	// Validate that the authorized application API key has enough rights to read and delete the device.
	sourceAppAccess, err := s.DCS.getApplicationAccess(sourceCtx, &sourceIDs.ApplicationIdentifiers)
	if err != nil {
		logger.WithError(err).Warn("Failed to get application access provider to verify authorized source application rights")
		return nil, err
	}
	sourceRights, err := sourceAppAccess.ListRights(sourceCtx, &sourceIDs.ApplicationIdentifiers, sourceCallOpts...)
	if err != nil {
		logger.WithError(err).Warn("Failed to list authorized source application rights")
		return nil, err
	}
	missingSourceRights := ttnpb.RightsFrom(
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
		ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
	).Sub(sourceRights).GetRights()
	if len(missingSourceRights) > 0 {
		logger.WithError(err).WithField("missing", missingSourceRights).Warn("Insufficient rights for source application")
		return nil, errPermissionDenied.New()
	}

	sourceCtx = events.ContextWithCorrelationID(sourceCtx, fmt.Sprintf("dcs:claim:%s", events.NewCorrelationID()))
	var sourceDev *ttnpb.EndDevice
	defer func() {
		if err == nil || sourceDev == nil {
			return
		}
		registerAbortClaimEndDevice(sourceCtx, sourceDev.EndDeviceIdentifiers, err)
	}()

	// Get source end device from Entity Registry.
	logger.Debug("Load source end device from Entity Registry")
	sourceDev, err = sourceERClient.Get(sourceCtx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIdentifiers: *sourceIDs,
		FieldMask: pbtypes.FieldMask{
			Paths: transferISPaths[:],
		},
	}, sourceCallOpts...)
	if err != nil {
		logger.WithError(err).Warn("Failed to load source end device from Entity Registry")
		return nil, err
	}
	logger = logger.WithField("join_server_address", sourceDev.JoinServerAddress)

	// Get source end device from Join Server.
	logger.Debug("Get source end device from Join Server")
	sourceJSClient, err := s.DCS.getJsDeviceRegistry(sourceCtx, sourceIDs)
	if err != nil {
		return nil, err
	}
	sourceJSDev, err := sourceJSClient.Get(sourceCtx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIdentifiers: *sourceIDs,
		FieldMask: pbtypes.FieldMask{
			Paths: transferJSPaths[:],
		},
	}, sourceCallOpts...)
	if err != nil {
		logger.WithError(err).Warn("Failed to get source end device from Join Server")
		return nil, err
	}
	if err := sourceDev.SetFields(sourceJSDev, ttnpb.ExcludeFields(transferJSPaths[:], "network_server_address", "application_server_address")...); err != nil {
		return nil, err
	}

	// Validate claim authentication code. Do not propagate the reason why the given authentication code is invalid.
	logger.Debug("Validate claim authentication code")
	if sourceDev.ClaimAuthenticationCode == nil {
		logger.Warn("Claim authentication code not specified")
		return nil, errClaimAuthenticationCode.WithAttributes("reason", "not_specified")
	}
	if sourceDev.ClaimAuthenticationCode.ValidFrom != nil && time.Since(*sourceDev.ClaimAuthenticationCode.ValidFrom) < 0 {
		logger.Warn("Claim authentication code not valid yet")
		return nil, errClaimAuthenticationCode.WithAttributes("reason", "too_early")
	}
	if sourceDev.ClaimAuthenticationCode.ValidTo != nil && time.Until(*sourceDev.ClaimAuthenticationCode.ValidTo) < 0 {
		logger.Warn("Claim authentication code not valid anymore")
		return nil, errClaimAuthenticationCode.WithAttributes("reason", "too_late")
	}
	if !strings.EqualFold(sourceDev.ClaimAuthenticationCode.Value, authCode) {
		logger.Warn("Claim authentication code mismatch")
		return nil, errClaimAuthenticationCode.WithAttributes("reason", "mismatch")
	}

	// Get source end device from Network Server and Application Server.
	var (
		skipTargetNSCreate bool
		sourceNSClient     ttnpb.NsEndDeviceRegistryClient
	)
	if sourceDev.NetworkServerAddress != "" {
		logger := logger.WithField("network_server_address", sourceDev.NetworkServerAddress)
		logger.Debug("Get source end device from Network Server")
		dialCtx, cancelDial := context.WithTimeout(sourceCtx, dialTimeout)
		defer cancelDial()
		sourceNSConn, err := discover.DialContext(dialCtx, sourceDev.NetworkServerAddress, credentials.NewTLS(sourceTLSConfig), sourceDialOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to dial source Network Server")
			return nil, err
		}
		defer sourceNSConn.Close()
		sourceNSClient = ttnpb.NewNsEndDeviceRegistryClient(sourceNSConn)
		sourceNSDev, err := sourceNSClient.Get(sourceCtx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: *sourceIDs,
			FieldMask: pbtypes.FieldMask{
				Paths: transferNSPaths[:],
			},
		}, sourceCallOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to get source end device from Network Server")
			if !errors.IsNotFound(err) {
				return nil, err
			}
			sourceNSClient = nil
			skipTargetNSCreate = true
		} else if err := sourceDev.SetFields(sourceNSDev, transferNSPaths[:]...); err != nil {
			return nil, err
		}
	}
	var (
		skipTargetASCreate bool
		sourceASClient     ttnpb.AsEndDeviceRegistryClient
	)
	if sourceDev.ApplicationServerAddress != "" {
		logger := logger.WithField("application_server_address", sourceDev.ApplicationServerAddress)
		logger.Debug("Get source end device from Application Server")
		dialCtx, cancelDial := context.WithTimeout(sourceCtx, dialTimeout)
		defer cancelDial()
		sourceASConn, err := discover.DialContext(dialCtx, sourceDev.ApplicationServerAddress, credentials.NewTLS(sourceTLSConfig), sourceDialOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to dial source Application Server")
			return nil, err
		}
		defer sourceASConn.Close()
		sourceASClient = ttnpb.NewAsEndDeviceRegistryClient(sourceASConn)
		sourceASDev, err := sourceASClient.Get(sourceCtx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: *sourceIDs,
			FieldMask: pbtypes.FieldMask{
				Paths: transferASPaths[:],
			},
		}, sourceCallOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to get source end device from Application Server")
			if !errors.IsNotFound(err) {
				return nil, err
			}
			sourceASClient = nil
			skipTargetASCreate = true
		} else if err := sourceDev.SetFields(sourceASDev, transferASPaths[:]...); err != nil {
			return nil, err
		}
	}

	// Before deleting the source end device, dial the target Network and Application Server to make sure they're
	// available.
	targetCallOpts := []grpc.CallOption{
		targetForwardAuth,
	}
	targetDialOpts := append([]grpc.DialOption(nil), rpcclient.DefaultDialOptions(targetCtx)...)
	targetDialOpts = append(targetDialOpts, grpc.WithBlock(), grpc.FailOnNonTempDialError(true))
	targetTLSConfig, err := s.DCS.GetTLSClientConfig(targetCtx)
	if err != nil {
		return nil, err
	}
	var targetNSConn *grpc.ClientConn
	if req.TargetNetworkServerAddress != "" {
		dialCtx, cancelDial := context.WithTimeout(targetCtx, dialTimeout)
		defer cancelDial()
		targetNSConn, err = discover.DialContext(dialCtx, req.TargetNetworkServerAddress, credentials.NewTLS(targetTLSConfig), targetDialOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to dial target Network Server")
			return nil, err
		}
		defer targetNSConn.Close()
	}
	var targetASConn *grpc.ClientConn
	if req.TargetApplicationServerAddress != "" {
		dialCtx, cancelDial := context.WithTimeout(targetCtx, dialTimeout)
		defer cancelDial()
		targetASConn, err = discover.DialContext(dialCtx, req.TargetApplicationServerAddress, credentials.NewTLS(targetTLSConfig), targetDialOpts...)
		if err != nil {
			logger.WithError(err).Warn("Failed to dial target Application Server")
			return nil, err
		}
		defer targetASConn.Close()
	}

	// Delete source end device in Application Server, Network Server, Join Server and Entity Registry, in ascending
	// order of importance. If any but not all deletions fails, the claiming process gets aborted and successfully deleted
	// devices are not recovered.
	for _, deleter := range []struct {
		name   string
		client interface {
			Delete(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
		}
		failOnError bool
	}{
		{
			name:        "Application Server",
			client:      sourceASClient,
			failOnError: sameHost(sourceDev.ApplicationServerAddress, req.TargetApplicationServerAddress),
		},
		{
			name:        "Network Server",
			client:      sourceNSClient,
			failOnError: sameHost(sourceDev.NetworkServerAddress, req.TargetNetworkServerAddress),
		},
		{
			name:        "Join Server",
			client:      sourceJSClient,
			failOnError: true,
		},
		{
			name:        "Entity Registry",
			client:      sourceERClient,
			failOnError: true,
		},
	} {
		if deleter.client == nil {
			continue
		}
		logger.Debugf("Delete source end device from %s", deleter.name)
		if _, err := deleter.client.Delete(sourceCtx, sourceIDs, sourceCallOpts...); err != nil {
			logger.WithError(err).Warnf("Failed to delete source end device from %s", deleter.name)
			if deleter.failOnError {
				return nil, err
			}
		}
	}

	defer func() {
		if err != nil {
			registerFailClaimEndDevice(sourceCtx, sourceDev, err)
		} else {
			registerSuccessClaimEndDevice(sourceCtx, sourceDev.EndDeviceIdentifiers)
		}
	}()

	targetDev := *sourceDev

	// Invalidate claim authentication code if requested.
	if req.InvalidateAuthenticationCode {
		now := time.Now()
		targetDev.ClaimAuthenticationCode = &ttnpb.EndDeviceAuthenticationCode{
			ValidFrom: sourceDev.ClaimAuthenticationCode.ValidFrom,
			ValidTo:   &now,
			Value:     sourceDev.ClaimAuthenticationCode.Value,
		}
	}

	// Set fields from the claiming request.
	targetDev.ApplicationIdentifiers = req.TargetApplicationIDs
	if req.TargetDeviceID != "" {
		targetDev.DeviceID = req.TargetDeviceID
	}
	logger = logger.WithField("target_device_uid", unique.ID(targetCtx, targetDev.EndDeviceIdentifiers))
	targetDev.NetworkServerAddress = req.TargetNetworkServerAddress
	targetDev.NetworkServerKEKLabel = req.TargetNetworkServerKEKLabel
	targetDev.ApplicationServerAddress = req.TargetApplicationServerAddress
	targetDev.ApplicationServerKEKLabel = req.TargetApplicationServerKEKLabel
	targetDev.ApplicationServerID = req.TargetApplicationServerID
	var targetISDev, targetJSDev, targetNSDev, targetASDev ttnpb.EndDevice
	for d, paths := range map[*ttnpb.EndDevice][]string{
		&targetISDev: transferISPaths[:],
		&targetJSDev: transferJSPaths[:],
		&targetNSDev: transferNSPaths[:],
		&targetASDev: transferASPaths[:],
	} {
		paths = append(paths, "ids")
		if err := d.SetFields(&targetDev, paths...); err != nil {
			return nil, err
		}
	}

	// Create target end device on Entity Registry and rollback if subsequent creates fail.
	logger.Debug("Create target end device on Entity Registry")
	targetERClient, err := s.DCS.getDeviceRegistry(targetCtx, &targetDev.EndDeviceIdentifiers)
	if err != nil {
		logger.WithError(err).Warn("Failed to get device registry")
		return nil, err
	}
	if _, err := targetERClient.Create(targetCtx, &ttnpb.CreateEndDeviceRequest{
		EndDevice: targetISDev,
	}, targetCallOpts...); err != nil {
		logger.WithError(err).Warn("Failed to create target end device on Entity Registry")
		return nil, err
	}
	defer func() {
		if err == nil {
			return
		}
		logger.WithError(err).Debug("Rollback create target end device on Entity Registry")
		if _, err := targetERClient.Delete(targetCtx, &targetDev.EndDeviceIdentifiers, targetCallOpts...); err != nil {
			logger.WithError(err).Warn("Failed to rollback create target end device on Entity Registry")
		}
	}()

	// Create target end device on Join Server and rollback if subsequent creates fail.
	logger.Debug("Create target end device on Join Server")
	targetJSClient, err := s.DCS.getJsDeviceRegistry(targetCtx, &targetDev.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	if _, err := targetJSClient.Set(targetCtx, &ttnpb.SetEndDeviceRequest{
		EndDevice: targetJSDev,
		FieldMask: pbtypes.FieldMask{
			Paths: transferJSPaths[:],
		},
	}, targetCallOpts...); err != nil {
		logger.WithError(err).Warn("Failed to create target end device on Join Server")
		return nil, err
	}
	defer func() {
		if err == nil {
			return
		}
		logger.WithError(err).Debug("Rollback create target end device on Join Server")
		if _, err := targetJSClient.Delete(targetCtx, &targetDev.EndDeviceIdentifiers, targetCallOpts...); err != nil {
			logger.WithError(err).Warn("Failed to rollback create target end device on Join Server")
		}
	}()

	// Create target end device on Network Server and rollback if subsequent creates fail.
	if targetNSConn != nil && !skipTargetNSCreate {
		logger := logger.WithField("network_server_address", targetDev.NetworkServerAddress)
		logger.Debug("Create target end device on Network Server")
		targetNSClient := ttnpb.NewNsEndDeviceRegistryClient(targetNSConn)
		if _, err := targetNSClient.Set(targetCtx, &ttnpb.SetEndDeviceRequest{
			EndDevice: targetNSDev,
			FieldMask: pbtypes.FieldMask{
				Paths: transferNSPaths[:],
			},
		}, targetCallOpts...); err != nil {
			logger.WithError(err).Warn("Failed to create target end device on Network Server")
			return nil, err
		}
		defer func() {
			if err == nil {
				return
			}
			logger.WithError(err).Debug("Rollback create target end device on Network Server")
			if _, err := targetNSClient.Delete(targetCtx, &targetDev.EndDeviceIdentifiers, targetCallOpts...); err != nil {
				logger.WithError(err).Warn("Failed to rollback create target end device on Network Server")
			}
		}()
	}

	// Create target end device on Application Server and rollback if subsequent creates fail.
	if targetASConn != nil && !skipTargetASCreate {
		logger := logger.WithField("application_server_address", targetDev.ApplicationServerAddress)
		logger.Debug("Create target end device on Application Server")
		targetASClient := ttnpb.NewAsEndDeviceRegistryClient(targetASConn)
		if _, err := targetASClient.Set(targetCtx, &ttnpb.SetEndDeviceRequest{
			EndDevice: targetASDev,
			FieldMask: pbtypes.FieldMask{
				Paths: transferASPaths[:],
			},
		}, targetCallOpts...); err != nil {
			logger.WithError(err).Warn("Failed to create target end device on Application Server")
			return nil, err
		}
		defer func() {
			if err == nil {
				return
			}
			logger.WithError(err).Debug("Rollback create target end device on Application Server")
			if _, err := targetASClient.Delete(targetCtx, &targetDev.EndDeviceIdentifiers, targetCallOpts...); err != nil {
				logger.WithError(err).Warn("Failed to rollback create target end device on Application Server")
			}
		}()
	}

	logger.Info("Claimed end device")
	return &targetDev.EndDeviceIdentifiers, nil
}

func (s *endDeviceClaimingServer) AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error) {
	_, err := s.DCS.authorizedAppsRegistry.Set(ctx, req.ApplicationIdentifiers, nil, func(key *ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error) {
		paths := make([]string, 0, 2)
		if key == nil {
			key = &ttipb.ApplicationAPIKey{
				ApplicationIDs: req.ApplicationIdentifiers,
			}
			paths = append(paths, "application_ids")
		}
		key.APIKey = req.APIKey
		paths = append(paths, "api_key")
		return key, paths, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *endDeviceClaimingServer) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	_, err := s.DCS.authorizedAppsRegistry.Set(ctx, *ids, nil, func(key *ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error) {
		if key == nil {
			return nil, nil, errAuthorizationNotFound.New()
		}
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
