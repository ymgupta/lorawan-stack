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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitButton from '../../../components/submit-button'
import SubmitBar from '../../../components/submit-bar'
import QR from '../../../components/qr'

import sharedMessages from '../../../lib/shared-messages'
import errorMessages from '../../../lib/errors/error-messages'
import PropTypes from '../../../lib/prop-types'
import { selectApplicationById } from '../../store/selectors/applications'
import { readQr } from '../../lib/qr'

import { deviceClaimValidationSchema } from './validation-schema'

import style from './device-claim-form.styl'

const m = defineMessages({
  claimRecognized: 'End device JoinEUI `{joinEUI}` and DevEUI `{devEUI}` recognized',
  claimAuthMessage: 'Scan authentication QR code',
})

@connect(function(state, props) {
  return {
    application: selectApplicationById(state, props.appId),
  }
})
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

    // return form inputs from application attributes.
    return Object.keys(attributes).map(key => {
      return <Form.Field title={key} name={key} type="text" component={Input} key={key} />
    })
  }

  handleSubmit = async (values, { setSubmitting }) => {
    /* eslint no-invalid-this: "off"*/
    const { onSubmit, onSubmitSuccess } = this.props
    const validationSchema = deviceClaimValidationSchema
    const castedValues = validationSchema.cast(values)
    await this.setState({ error: '' })

    try {
      const device = await onSubmit(castedValues)
      await onSubmitSuccess(device)
    } catch (error) {
      setSubmitting(false)
      const err = error instanceof Error ? errorMessages.genericError : error
      await this.setState({ error: err, info: '' })
    }
  }

  @bind
  handleQrCodeChange(qrCode) {
    const { devEUI, joinEUI } = readQr(qrCode)
    const {
      application: { attributes },
    } = this.props

    return !attributes && qrCode
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
          validationSchema={deviceClaimValidationSchema}
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
