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
import { defineMessages } from 'react-intl'
import PropTypes from '../../../lib/prop-types'
import api from '../../api'

import sharedMessages from '../../../lib/shared-messages'
import PageTitle from '../../../components/page-title'

import DeviceClaimForm from '../../containers/device-claim-form'

import style from './device-claim.styl'

const m = defineMessages({
  claimSuccess: 'End device JoinEUI `{joinEUI}` and DevEUI `{devEUI}` claimed',
})

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
    const { qrCode, ...attributes } = values
    const claim = await api.deviceClaim.claim(appId, qrCode)
    return await api.device.update(appId, claim.device_id, {
      attributes,
    })
  }

  handleSubmitSuccess = device => {
    /* eslint no-invalid-this: "off"*/
    const { redirectHome } = this.props
    const { dev_eui: devEUI, join_eui: joinEUI } = device.ids
    redirectHome({
      values: { devEUI, joinEUI },
      ...m.claimSuccess,
    })
  }

  render() {
    const { appId } = this.props
    return (
      <Container>
        <PageTitle title={sharedMessages.claimDevice} className={style.title} />
        <Row>
          <Col>
            <DeviceClaimForm
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              appId={appId}
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
