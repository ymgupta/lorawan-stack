// Copyright © 2020 The Things Industries B.V.

import React, { Component } from 'react'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'
import Yup from '@ttn-lw/lib/yup'
import { merge } from 'lodash'

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

export const blankValues = {
  region: '',
  access_key: {
    access_key_id: '',
    secret_access_key: '',
    session_token: '',
  },
  _use_access_key: false,
  assume_role: {
    arn: '',
    external_id: '',
    session_duration: '',
  },
  _use_assume_role: false,
  endpoint_address: '',
  default: {
    stack_name: '',
  },
  _use_default: true,
}

export const mapToFormValues = function(awsIoT) {
  return merge(
    {
      _use_access_key: 'access_key' in awsIoT,
      _use_assume_role: 'assume_role' in awsIoT,
      _use_default: 'default' in awsIoT,
    },
    blankValues,
    awsIoT,
  )
}

export const mapFromFormValues = function(result) {
  result.format = 'json'
  delete result.aws_iot._use_access_key
  delete result.aws_iot._use_assume_role
  delete result.aws_iot._use_default
  return result
}

const regions = {
  'ap-northeast-1': 'Asia Pacific, Tokyo',
  'ap-northeast-2': 'Asia Pacific, Seoul',
  'ap-south-1': 'Asia Pacific, Mumbai',
  'ap-southeast-1': 'Asia Pacific, Singapore',
  'ap-southeast-2': 'Asia Pacific, Sydney',
  'ca-central-1': 'Canada',
  'eu-central-1': 'Europe, Frankfurt',
  'eu-north-1': 'Europe, Stockholm',
  'eu-west-1': 'Europe, Ireland',
  'eu-west-2': 'Europe, London',
  'eu-west-3': 'Europe, Paris',
  'sa-east-1': 'South America, São Paulo',
  'us-east-1': 'United States, North Virginia',
  'us-east-2': 'United States, Ohio',
  'us-west-1': 'United States, North California',
  'us-west-2': 'United States, Oregon',
}

const endpointAddressRegexp = /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])$/
const accessKeyIDRegexp = /^[\w]*$/
const assumeRoleArnRegexp = /^arn:aws:iam::[0-9]{12}:role\/[A-Za-z0-9_+=,.@-]+$/
const externalIDRegexp = /^[\w+=,.@:/-]*$/
const stackNameRegexp = /^[A-Za-z][A-Za-z0-9-]*$/
const sessionDurationRegexp = /^[0-9]{1,}[.]?([0-9]{1,})?[a-zA-Z]{1,2}$/

const m = defineMessages({
  config: 'AWS IoT configuration',
  useDefault: 'Use default integration',
  defaultInfo:
    'The default AWS IoT integration can be deployed via CloudFormation in your AWS account. {moreInformation}',
  region: 'Region',
  endpointAddress: 'Endpoint address',
  endpointAddressDescription:
    'If the endpoint address is left empty, the integration will try to discover it.',
  accessKey: 'Access key',
  accessKeyID: 'Access key ID',
  accessKeySecret: 'Secret access key',
  accessKeySessionToken: 'Session token',
  assumeRole: 'Assume role',
  assumeRoleArn: 'Role ARN',
  assumeRoleExternalID: 'External ID',
  assumeRoleSessionDuration: 'Session duration',
  defaultStackName: 'CloudFormation stack name',
  defaultStackNameDescription:
    'Copy the CloudFormation stack name that you used when deploying the integration in your AWS account.',
  defaultRoleArn: 'Cross-account role ARN',
  defaultRoleArnDescription: 'Copy the CloudFormation stack "CrossAccountRoleArn" output.',
  validateAccessKeyIDFormat: '{field} must be a valid access key ID',
  validateEndpointAddressFormat: '{field} must be a valid AWS IoT endpoint address',
  validateRoleArnFormat: '{field} must be a valid role ARN',
  validateExternalIDFormat: '{field} must be a valid external ID',
  validateSessionDurationFormat: '{field} must be a valid session duration',
  validateStackNameFormat: '{field} is not a valid stack name',
})

export const validationSchema = Yup.object().shape({
  region: Yup.string()
    .oneOf(Object.keys(regions))
    .required(sharedMessages.validateRequired),
  access_key: Yup.object().when('_use_access_key', {
    is: true,
    then: Yup.object().shape({
      access_key_id: Yup.string()
        .matches(accessKeyIDRegexp, m.validateAccessKeyIDFormat)
        .min(16, Yup.passValues(sharedMessages.validateTooShort))
        .max(128, Yup.passValues(sharedMessages.validateTooLong))
        .required(sharedMessages.validateRequired),
      secret_access_key: Yup.string()
        .max(40, Yup.passValues(sharedMessages.validateTooLong))
        .required(sharedMessages.validateRequired),
      session_token: Yup.string().max(256, Yup.passValues(sharedMessages.validateTooLong)),
    }),
    otherwise: Yup.object().strip(),
  }),
  assume_role: Yup.object().when(
    ['_use_assume_role', '_use_default'],
    (useAssumeRole, useDefault, schema) => {
      if (useDefault) {
        return schema.shape({
          arn: Yup.string()
            .matches(assumeRoleArnRegexp, m.validateRoleArnFormat)
            .required(sharedMessages.validateRequired),
          external_id: Yup.string().strip(),
          session_duration: Yup.string().strip(),
        })
      }
      if (useAssumeRole) {
        return schema.shape({
          arn: Yup.string()
            .matches(assumeRoleArnRegexp, m.validateRoleArnFormat)
            .required(sharedMessages.validateRequired),
          external_id: Yup.string()
            .matches(externalIDRegexp, m.validateExternalIDFormat)
            .min(2, Yup.passValues(sharedMessages.validateTooShort))
            .max(1224, Yup.passValues(sharedMessages.validateTooLong)),
          session_duration: Yup.string().matches(
            sessionDurationRegexp,
            m.validateSessionDurationFormat,
          ),
        })
      }
      return schema.strip()
    },
  ),
  endpoint_address: Yup.string().matches(endpointAddressRegexp, m.validateEndpointAddressFormat),
  default: Yup.object().when('_use_default', {
    is: true,
    then: Yup.object().shape({
      stack_name: Yup.string()
        .matches(stackNameRegexp, m.validateStackNameFormat)
        .required(sharedMessages.validateRequired),
    }),
    otherwise: Yup.object().strip(),
  }),
})

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
