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
import { connect } from 'react-redux'
import * as Yup from 'yup'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import QR from '@ttn-lw/components/qr'
import withRequest from '@ttn-lw/lib/components/with-request'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import errorMessages from '@ttn-lw/lib/errors/error-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { readQr } from '@claim/lib/qr'
import { getApplication } from '@claim/store/actions/applications'
import {
  selectApplicationById,
  selectApplicationFetching,
} from '@claim/store/selectors/applications'

import style from './device-claim-form.styl'

const m = defineMessages({
  claimRecognized: 'End device JoinEUI `{joinEUI}` and DevEUI `{devEUI}` recognized',
  claimAuthMessage: 'Scan authentication QR code',
  claimInvalid: 'Invalid authentication QR code',
})

const validationSchema = Yup.object().shape({
  qrCode: Yup.string()
    .matches(/^URN:DEV:LW/, m.claimInvalid)
    .required(m.claimInvalid),
})

@connect(
  function(state, props) {
    return {
      fetching: selectApplicationFetching(state),
      application: selectApplicationById(state, props.appId),
    }
  },
  dispatch => ({
    loadData: id => {
      dispatch(getApplication(id, ['name,attributes']))
    },
  }),
)
@withRequest(
  ({ appId, loadData }) => loadData(appId),
  ({ fetching, application }) => fetching || !Boolean(application),
)
export default class DeviceClaimForm extends Component {
  static propTypes = {
    application: PropTypes.application,
    onSubmit: PropTypes.func.isRequired,
    onSubmitSuccess: PropTypes.func,
  }

  static defaultProps = {
    onSubmitSuccess: () => null,
    application: {},
  }

  formRef = React.createRef()
  state = {
    error: '',
    info: '',
    isQrCode: false,
  }

  addApplicationAttributes() {
    const {
      application: { attributes },
    } = this.props

    // Return form inputs from application attributes.
    return Object.keys(attributes).map(key => {
      return <Form.Field title={key} name={key} type="text" component={Input} key={key} />
    })
  }

  handleSubmit = async (values, { setSubmitting }) => {
    /* eslint no-invalid-this: "off"*/
    const { onSubmit, onSubmitSuccess } = this.props
    await this.setState({ error: '' })

    try {
      const device = await onSubmit(values)
      await onSubmitSuccess(device)
    } catch (error) {
      setSubmitting(false)
      const err = error instanceof Error ? errorMessages.genericError : error
      await this.setState({ error: err, info: '' })
    }
  }

  @bind
  async handleQrCodeChange(qrCode) {
    const valid = await validationSchema.isValid({ qrCode })
    if (!valid) return

    const { devEUI, joinEUI } = readQr(qrCode)
    const {
      application: { attributes },
    } = this.props

    return !attributes
      ? this.formRef.current.submitForm()
      : this.setState({
          isQrCode: true,
          info: {
            values: { devEUI, joinEUI },
            ...m.claimRecognized,
          },
        })
  }

  render() {
    const { error, isQrCode, info } = this.state
    const initialValues = {}

    return (
      <div>
        <Form
          error={error}
          info={info}
          onSubmit={this.handleSubmit}
          initialValues={initialValues}
          validationSchema={validationSchema}
          validateOnChange
          formikRef={this.formRef}
        >
          {isQrCode ? (
            <>
              {this.addApplicationAttributes()}
              <SubmitBar>
                <Form.Submit component={SubmitButton} message={sharedMessages.claimDevice} />
              </SubmitBar>
            </>
          ) : (
            <Form.Field
              title={sharedMessages.claimAuth}
              name="qrCode"
              description={m.claimAuthMessage}
              component={QR}
              className={style.qrField}
              onChange={this.handleQrCodeChange}
            />
          )}
        </Form>
      </div>
    )
  }
}
