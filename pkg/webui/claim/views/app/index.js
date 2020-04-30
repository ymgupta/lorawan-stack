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
import { ConnectedRouter } from 'connected-react-router'

import { Route, Switch } from 'react-router-dom'

import Footer from '@ttn-lw/components/footer'
import { ToastContainer } from '@ttn-lw/components/toast'
import ErrorView from '@ttn-lw/lib/components/error-view'
import { withEnv } from '@ttn-lw/lib/components/env'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Header from '@claim/containers/header'
import Landing from '@claim/views/landing'
import Login from '@claim/views/login'
import FullViewError from '@claim/views/error'
import dev from '@ttn-lw/lib/dev'

import style from './app.styl'

@withEnv
class ClaimApp extends React.Component {
  render() {
    const {
      env: { siteTitle, pageData, siteName },
      history,
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
              <div className={style.content}>
                <Switch>
                  {/* Routes for registration, privacy policy, other public pages. */}
                  <Route path="/login" component={Login} />
                  <Route path="/" component={Landing} />
                </Switch>
              </div>
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
