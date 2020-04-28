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

import api from '@claim/api'
import * as applications from '@claim/store/actions/applications'

import createRequestLogic from './lib'

const getApplicationLogic = createRequestLogic({
  type: applications.GET_APP,
  async process({ action }, dispatch) {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const app = await api.application.get(id, selector)
    return app
  },
})

const updateApplicationLogic = createRequestLogic({
  type: applications.UPDATE_APP,
  async process({ action }) {
    const { id, patch } = action.payload

    const result = await api.application.update(id, patch)

    return { ...patch, ...result }
  },
})

const deleteApplicationLogic = createRequestLogic({
  type: applications.DELETE_APP,
  async process({ action }) {
    const { id } = action.payload

    await api.application.delete(id)

    return { id }
  },
})

const getApplicationsLogic = createRequestLogic({
  type: applications.GET_APPS_LIST,
  latest: true,
  async process({ action }) {
    const {
      params: { page, limit, query },
    } = action.payload
    const { selectors } = action.meta

    const data = query
      ? await api.applications.search(
          {
            page,
            limit,
            id_contains: query,
            name_contains: query,
          },
          selectors,
        )
      : await api.applications.list({ page, limit }, selectors)

    return { entities: data.applications, totalCount: data.totalCount }
  },
})

export default [
  getApplicationLogic,
  updateApplicationLogic,
  deleteApplicationLogic,
  getApplicationsLogic,
]
