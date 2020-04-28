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

import axios from 'axios'
import TTN from 'ttn-lw'

import getCookieValue from '@ttn-lw/lib/cookie'
import { selectStackConfig, selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'
import token from '@claim/lib/access-token'

const stackConfig = selectStackConfig()
const appRoot = selectApplicationRootPath()

const stack = {
  ns: stackConfig.ns.enabled ? stackConfig.ns.base_url : undefined,
  as: stackConfig.as.enabled ? stackConfig.as.base_url : undefined,
  is: stackConfig.is.enabled ? stackConfig.is.base_url : undefined,
  dcs: stackConfig.dcs.enabled ? stackConfig.dcs.base_url : undefined,
}

const isBaseUrl = stackConfig.is.base_url

const ttnClient = new TTN(token, {
  stackConfig: stack,
  connectionType: 'http',
  proxy: false,
})

const csrf = getCookieValue('_csrf')
const instance = axios.create({
  headers: { 'X-CSRF-Token': csrf },
})

export default {
  claim: {
    token() {
      return instance.get(`${appRoot}/api/auth/token`)
    },
    async logout() {
      let csrf = getCookieValue('_claim_csrf')

      if (!csrf) {
        // If the csrf token has been deleted, we likely have some outside
        // manipulation of the cookies. We can try to regain the cookie by
        // making an AJAX request to the current location.
        await axios.get(window.location)
        csrf = getCookieValue('_claim_csrf')

        if (!csrf) {
          // If we still could not retrieve the cookie, throw an error.
          throw new Error('Could not retrieve the csrf token')
        }
      }

      return instance.post(`${appRoot}/api/auth/logout`, undefined, {
        headers: { 'X-CSRF-Token': csrf },
      })
    },
  },
  clients: {
    get(client_id) {
      return instance.get(`${isBaseUrl}/is/clients/${client_id}`)
    },
  },
  users: {
    async get(userId) {
      return instance.get(`${isBaseUrl}/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${(await token()).access_token}`,
        },
      })
    },
    async authInfo() {
      return instance.get(`${isBaseUrl}/auth_info`, {
        headers: {
          Authorization: `Bearer ${(await token()).access_token}`,
        },
      })
    },
  },
  deviceClaim: {
    claim: ttnClient.DeviceClaim.claim.bind(ttnClient.DeviceClaim),
  },
  device: {
    update: ttnClient.Applications.Devices.updateById.bind(ttnClient.Applications.Devices),
  },
  applications: {
    list: ttnClient.Applications.getAll.bind(ttnClient.Applications),
  },
  application: {
    get: ttnClient.Applications.getById.bind(ttnClient.Applications),
  },
}
