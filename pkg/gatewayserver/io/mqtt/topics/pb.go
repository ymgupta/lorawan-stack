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

package topics

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const topicV3 = "v3"

type v3 struct {
	multitenancy bool
}

func (v3 *v3) BirthTopic(uid string) []string {
	return nil
}

func (v3 *v3) IsBirthTopic(path []string) bool {
	return false
}

func (v3 *v3) LastWillTopic(uid string) []string {
	return nil
}

func (v3 *v3) IsLastWillTopic(path []string) bool {
	return false
}

func (v3 *v3) UplinkTopic(uid string) []string {
	if !v3.multitenancy {
		ids, _ := unique.ToGatewayID(uid) // The error can be safely ignored here since the caller already validates the uid.
		return []string{topicV3, ids.GatewayID, "up"}
	}
	return []string{topicV3, uid, "up"}
}

func (v3 *v3) IsUplinkTopic(path []string) bool {
	return len(path) == 3 && path[0] == topicV3 && path[2] == "up"
}

func (v3 *v3) StatusTopic(uid string) []string {
	if !v3.multitenancy {
		ids, _ := unique.ToGatewayID(uid)
		return []string{topicV3, ids.GatewayID, "status"}
	}
	return []string{topicV3, uid, "status"}
}

func (v3 *v3) IsStatusTopic(path []string) bool {
	return len(path) == 3 && path[0] == topicV3 && path[2] == "status"
}

func (v3 *v3) TxAckTopic(uid string) []string {
	if !v3.multitenancy {
		ids, _ := unique.ToGatewayID(uid)
		return []string{topicV3, ids.GatewayID, "down", "ack"}
	}
	return []string{topicV3, uid, "down", "ack"}
}

func (v3 *v3) IsTxAckTopic(path []string) bool {
	return len(path) == 4 && path[0] == topicV3 && path[2] == "down" && path[3] == "ack"
}

func (v3 *v3) DownlinkTopic(uid string) []string {
	if !v3.multitenancy {
		ids, _ := unique.ToGatewayID(uid)
		return []string{topicV3, ids.GatewayID, "down"}
	}
	return []string{topicV3, uid, "down"}
}

// New returns the default layout.
func New(ctx context.Context) Layout {
	if license.RequireMultiTenancy(ctx) == nil {
		return &v3{
			multitenancy: true,
		}
	}
	return &v3{}
}
