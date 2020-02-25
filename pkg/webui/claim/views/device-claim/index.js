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
import PageTitle from '../../../components/page-title'

import DeviceClaimForm from '../../containers/device-claim-form'

import style from './device-claim.styl'

@connect(
  function(state, props) {
    return {
      appId: props.match.params.appId,
    }
  },
  {
    redirectHome: message =>
      push('/', {
        message,
      }),
  },
)
export default class DeviceClaim extends Component {
  handleSubmit = async values => {
    const { appId } = this.props
    try {
      const device = await api.deviceClaim.claim(values.qrCode, appId)
      return device
    } catch (err) {
      throw err
    }
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
        <PageTitle title={sharedMessages.claimDevice} className={style.title} />
        <Row>
          <Col>
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
  redirectHome: () => null,
}
