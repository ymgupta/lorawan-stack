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

import { combineReducers } from 'redux'
import { connectRouter } from 'connected-react-router'

import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import { SHARED_NAME_SINGLE as APPLICATION_SHARED_NAME } from '@claim/store/actions/applications'
import user from './user'
import init from './init'
import applications from './applications'
import fetching from './ui/fetching'
import error from './ui/error'
import { createNamedPaginationReducer } from './pagination'

export default history =>
  combineReducers({
    user,
    init,
    applications,
    ui: combineReducers({
      fetching,
      error,
    }),
    pagination: combineReducers({
      applications: createNamedPaginationReducer(APPLICATION_SHARED_NAME, getApplicationId),
    }),
    router: connectRouter(history),
  })
