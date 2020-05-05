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

import { hot } from 'react-hot-loader/root'
import React from 'react'
import { connect } from 'react-redux'
import { ConnectedRouter } from 'connected-react-router'

import { Route, Switch } from 'react-router-dom'

import Footer from '@ttn-lw/components/footer'
import { ToastContainer } from '@ttn-lw/components/toast'
import ErrorView from '@ttn-lw/lib/components/error-view'
import { withEnv } from '@ttn-lw/lib/components/env'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import WithAuth from '@ttn-lw/lib/components/with-auth'
import Header from '@claim/containers/header'
import Overview from '@claim/views/overview'
import DeviceClaim from '@claim/views/device-claim'
import FullViewError, { FullViewErrorInner } from '@claim/views/error'
import dev from '@ttn-lw/lib/dev'

import PropTypes from '@ttn-lw/lib/prop-types'
import {
  selectUser,
  selectUserFetching,
  selectUserError,
  selectUserRights,
  selectUserIsAdmin,
} from '@claim/store/selectors/user'

import style from './app.styl'

const GenericNotFound = () => <FullViewErrorInner error={{ statusCode: 404 }} />

@withEnv
@connect(state => ({
  user: selectUser(state),
  fetching: selectUserFetching(state),
  error: selectUserError(state),
  rights: selectUserRights(state),
  isAdmin: selectUserIsAdmin(state),
}))
class ClaimApp extends React.Component {
  static propTypes = {
    env: PropTypes.env.isRequired,
    error: PropTypes.error,
    fetching: PropTypes.bool.isRequired,
    history: PropTypes.shape({
      push: PropTypes.func,
      replace: PropTypes.func,
    }).isRequired,
    isAdmin: PropTypes.bool,
    rights: PropTypes.rights,
    user: PropTypes.user,
  }
  static defaultProps = {
    user: undefined,
    error: undefined,
    isAdmin: undefined,
    rights: undefined,
  }

  render() {
    const {
      history,
      user,
      fetching,
      error,
      rights,
      isAdmin,
      env: { siteTitle, pageData, siteName },
    } = this.props

    if (pageData && pageData.error) {
      return (
        <ConnectedRouter history={history}>
          <FullViewError error={pageData.error} />
        </ConnectedRouter>
      )
    }

    return (
      <ConnectedRouter history={history}>
        <ErrorView ErrorComponent={FullViewError}>
          <div className={style.app}>
            <IntlHelmet
              titleTemplate={`%s - ${siteTitle ? `${siteName} - ` : ''}${siteName}`}
              defaultTitle={siteName}
            />
            <div id="modal-container" />
            <Header className={style.header} />
            <main className={style.main}>
              <WithAuth
                user={user}
                fetching={fetching}
                error={error}
                errorComponent={FullViewErrorInner}
                rights={rights}
                isAdmin={isAdmin}
              >
                <div className={style.content}>
                  <Switch>
                    {/* Routes for registration, privacy policy, other public pages. */}
                    <Route exact path="/" component={Overview} />
                    <Route path={'/:appId'} component={DeviceClaim} />
                    <Route component={GenericNotFound} />
                  </Switch>
                </div>
              </WithAuth>
            </main>
            <ToastContainer />
            <Footer className={style.footer} />
          </div>
        </ErrorView>
      </ConnectedRouter>
    )
  }
}

const ExportedApp = dev ? hot(ClaimApp) : ClaimApp

export default ExportedApp
