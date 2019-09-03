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

import React, { Component } from 'react'
import { Row, Col, Container } from 'react-grid-system'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import PropTypes from '../../../lib/prop-types'
import api from '../../api'

import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'
import withRequest from '../../../lib/components/with-request'

import DeviceClaimForm from '../../containers/device-claim-form'

import { getApplication } from '../../store/actions/applications'
import {
  selectSelectedApplication,
  selectApplicationFetching,
  selectApplicationError,
} from '../../store/selectors/applications'

@connect(
  function(state, props) {
    return {
      appId: props.match.params.appId,
      fetching: selectApplicationFetching(state),
      application: selectSelectedApplication(state),
      error: selectApplicationError(state),
    }
  },
  dispatch => ({
    getApplication: id => dispatch(getApplication(id, 'name,description')),
    redirectHome: message =>
      dispatch(
        push('/', {
          message,
        }),
      ),
  }),
)
@withRequest(
  ({ appId, getApplication }) => getApplication(appId),
  ({ fetching, application }) => fetching || !Boolean(application),
)
export default class DeviceClaim extends Component {
  handleSubmit = async values => {
    const { appId } = this.props
    // Example QR code: URN:LW:DP:0016C00000000001:58A0CBFFFFFEFFFF:FACE800C:%VF1FB552B%S201907160010
    return api.device.claim({
      qr_code: btoa(values.claim_id),
      target_application_ids: {
        application_id: appId,
      },
    })
  }

  handleSubmitSuccess = () => {
    /* eslint no-invalid-this: "off"*/
    const { redirectHome } = this.props
    const message = sharedMessages.claimSuccess
    redirectHome(message)
  }

  render() {
    return (
      <Container>
        <IntlHelmet title={sharedMessages.claimDevice} />
        <Row>
          <Col sm={12}>
            <Message component="h2" content={sharedMessages.claimDevice} />
          </Col>
          <Col sm={12}>
            <DeviceClaimForm
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}

DeviceClaim.propTypes = {
  appId: PropTypes.string,
  redirectHome: PropTypes.func,
}

DeviceClaim.defaultProps = {
  appId: '',
  redirectHome() {
    return false
  },
}
