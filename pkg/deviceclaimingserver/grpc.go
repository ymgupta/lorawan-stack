// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deviceclaimingserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type endDeviceClaimingServer struct {
	DCS *DeviceClaimingServer
}

func (s *endDeviceClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (*ttnpb.EndDeviceIdentifiers, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
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

var errApplicationNotAuthorized = errors.DefineNotFound("application_not_authorized", "application not authorized")

func (s *endDeviceClaimingServer) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	_, err := s.DCS.authorizedAppsRegistry.Set(ctx, *ids, nil, func(key *ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error) {
		if key == nil {
			return nil, nil, errApplicationNotAuthorized
		}
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
