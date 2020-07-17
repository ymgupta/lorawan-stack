// Copyright © 2020 The Things Industries B.V.

import React, { Component } from 'react'
import bind from 'autobind-decorator'

import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Link from '@ttn-lw/components/link'
import Notification from '@ttn-lw/components/notification'
import Select from '@ttn-lw/components/select'
import UnitInput from '@ttn-lw/components/unit-input'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { unit as unitRegexp, emptyDuration as emptyDurationRegexp } from '@console/lib/regexp'

import m from './messages.tti'
import regions from './regions.tti'

export default class AWSIoTSettings extends Component {
  static propTypes = {
    onUseDefaultsChange: PropTypes.func,
    value: PropTypes.pubsubAWSIoT.isRequired,
  }

  static defaultProps = {
    onUseDefaultsChange: () => null,
  }

  constructor(props) {
    super(props)

    this.state = {
      useAccessKey: props.value._use_access_key,
      useAssumeRole: props.value._use_assume_role,
      useDefault: props.value._use_default,
    }
  }

  @bind
  handleUseAccessKeyChange(event) {
    this.setState({ useAccessKey: event.target.checked })
  }

  @bind
  handleUseAssumeRoleChange(event) {
    this.setState({ useAssumeRole: event.target.checked })
  }

  @bind
  handleUseDefaultChange(event) {
    this.setState({ useDefault: event.target.checked })
    this.props.onUseDefaultsChange(event.target.checked)
  }

  decodeSessionDurationValue(value) {
    if (emptyDurationRegexp.test(value)) {
      return {
        duration: undefined,
        unit: value,
      }
    }
    const duration = value.split(unitRegexp)[0]
    const unit = value.split(duration)[1]
    return {
      duration: duration ? Number(duration) : undefined,
      unit,
    }
  }

  render() {
    const { useAccessKey, useAssumeRole, useDefault } = this.state

    return (
      <React.Fragment>
        <Form.SubTitle title={m.config} />
        <Notification
          info
          small
          content={m.defaultInfo}
          messageValues={{
            moreInformation: (
              <Link.DocLink path="/integrations/pubsub/aws-iot/">
                <Message content={sharedMessages.moreInformation} />
              </Link.DocLink>
            ),
          }}
        />
        <Form.Field
          name="aws_iot._use_default"
          title={m.useDefault}
          component={Checkbox}
          onChange={this.handleUseDefaultChange}
        />
        <Form.Field
          name="aws_iot.region"
          title={m.region}
          component={Select}
          options={Object.keys(regions)
            .sort()
            .map(region => ({ value: region, label: `${region} (${regions[region]})` }))}
        />
        {useDefault && (
          <React.Fragment>
            <Form.Field
              name="aws_iot.default.stack_name"
              title={m.defaultStackName}
              component={Input}
            />
            <Form.Field
              name="aws_iot.assume_role.arn"
              title={m.defaultRoleArn}
              component={Input}
              description={m.defaultRoleArnDescription}
            />
          </React.Fragment>
        )}
        {!useDefault && (
          <React.Fragment>
            <Form.Field
              name="aws_iot.endpoint_address"
              title={m.endpointAddress}
              component={Input}
              description={m.endpointAddressDescription}
            />
            <Form.SubTitle title={m.accessKey} />
            <Form.Field
              name="aws_iot._use_access_key"
              title={m.accessKey}
              component={Checkbox}
              onChange={this.handleUseAccessKeyChange}
            />
            <Form.Field
              name="aws_iot.access_key.access_key_id"
              title={m.accessKeyID}
              component={Input}
              required={useAccessKey}
              disabled={!useAccessKey}
            />
            <Form.Field
              name="aws_iot.access_key.secret_access_key"
              title={m.accessKeySecret}
              component={Input}
              disabled={!useAccessKey}
            />
            <Form.Field
              name="aws_iot.access_key.session_token"
              title={m.accessKeySessionToken}
              component={Input}
              disabled={!useAccessKey}
            />
            <Form.SubTitle title={m.assumeRole} />
            <Form.Field
              name="aws_iot._use_assume_role"
              title={m.assumeRole}
              component={Checkbox}
              onChange={this.handleUseAssumeRoleChange}
            />
            <Form.Field
              name="aws_iot.assume_role.arn"
              title={m.assumeRoleArn}
              component={Input}
              required={useAssumeRole}
              disabled={!useAssumeRole}
            />
            <Form.Field
              name="aws_iot.assume_role.external_id"
              title={m.assumeRoleExternalID}
              component={Input}
              disabled={!useAssumeRole}
            />
            <Form.Field
              name="aws_iot.assume_role.session_duration"
              title={m.assumeRoleSessionDuration}
              component={UnitInput}
              units={[
                { label: sharedMessages.minutes, value: 'm' },
                { label: sharedMessages.hours, value: 'h' },
              ]}
              decode={this.decodeSessionDurationValue}
              disabled={!useAssumeRole}
            />
          </React.Fragment>
        )}
      </React.Fragment>
    )
  }
}
