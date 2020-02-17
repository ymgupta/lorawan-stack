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

import Form from '../../../components/form'
import SubmitButton from '../../../components/submit-button'
import SubmitBar from '../../../components/submit-bar'
import QR from '../../../components/qr'

import sharedMessages from '../../../lib/shared-messages'
import errorMessages from '../../../lib/errors/error-messages'
import PropTypes from '../../../lib/prop-types'
import m from './messages'

import { deviceClaimValidationSchema } from './validation-schema'

import style from './device-claim-form.styl'

export default class DeviceClaimForm extends Component {
  constructor(props) {
    super(props)
    this.state = { error: '' } // state needs to be defined to be used in render function.
  }

  handleSubmit = async (values, { setSubmitting }) => {
    /* eslint no-invalid-this: "off"*/
    const { onSubmit, onSubmitSuccess } = this.props
    const validationSchema = deviceClaimValidationSchema
    const castedValues = validationSchema.cast(values)
    await this.setState({ error: '' })

    try {
      await onSubmit(castedValues)
      await onSubmitSuccess()
    } catch (error) {
      setSubmitting(false)
      const err = error instanceof Error ? errorMessages.genericError : error
      await this.setState({ error: err })
    }
  }

  render() {
    const { error } = this.state
    const initialValues = {}

    return (
      <div>
        <Form
          error={error}
          onSubmit={this.handleSubmit}
          initialValues={initialValues}
          validationSchema={deviceClaimValidationSchema}
        >
          <Form.Field
            title={sharedMessages.claimAuth}
            name="qrCode"
            description={m.ClaimAuthMessage}
            component={QR}
            className={style.qrField}
          />
          <SubmitBar>
            <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
          </SubmitBar>
        </Form>
      </div>
    )
  }
}

DeviceClaimForm.propTypes = {
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func,
}

DeviceClaimForm.defaultProps = {
  onSubmitSuccess: () => null,
}
