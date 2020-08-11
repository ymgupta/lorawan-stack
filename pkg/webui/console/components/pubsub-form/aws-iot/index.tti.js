// Copyright © 2020 The Things Industries B.V.

import React from 'react'

import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Link from '@ttn-lw/components/link'
import Notification from '@ttn-lw/components/notification'
import Select from '@ttn-lw/components/select'
import UnitInput from '@ttn-lw/components/unit-input'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import m from './messages.tti'
import regions from './regions.tti'

const AWSIoTSettings = props => {
  const { onUseDefaultsChange, value } = props

  const [useAccessKey, setUseAccessKey] = React.useState(value._use_access_key)
  const handleUseAccessKeyChange = React.useCallback(() => {
    setUseAccessKey(!useAccessKey)
  }, [setUseAccessKey, useAccessKey])

  const [useAssumeRole, setUseAssumeRole] = React.useState(value._use_assume_role)
  const handleUseAssumeRoleChange = React.useCallback(() => {
    setUseAssumeRole(!useAssumeRole)
  }, [setUseAssumeRole, useAssumeRole])

  const [useDefault, setUseDefault] = React.useState(value._use_default)
  const handleUseDefaultChange = React.useCallback(() => {
    setUseDefault(!useDefault)
    onUseDefaultsChange(!useDefault)
  }, [onUseDefaultsChange, setUseDefault, useDefault])

  return (
    <>
      <Form.SubTitle title={m.config} />
      <Notification
        info
        small
        content={m.defaultInfo}
        messageValues={{
          awsIoTDoc: (
            <Link.DocLink primary path="/integrations/aws-iot/default/">
              AWS IoT
            </Link.DocLink>
          ),
        }}
      />
      <Form.Field
        name="aws_iot._use_default"
        title={m.useDefault}
        component={Checkbox}
        onChange={handleUseDefaultChange}
      />
      <Form.Field
        name="aws_iot.region"
        title={m.region}
        component={Select}
        options={Object.keys(regions)
          .map(region => ({ value: region, label: `${regions[region]} (${region})` }))
          .sort((a, b) => a.label.localeCompare(b.label))}
      />
      {useDefault && (
        <>
          <Form.Field
            name="aws_iot.default.stack_name"
            title={m.defaultStackName}
            component={Input}
            required
          />
          <Form.Field
            name="aws_iot.assume_role.arn"
            title={m.defaultRoleArn}
            component={Input}
            description={m.defaultRoleArnDescription}
            placeholder={m.assumeRoleArnPlaceholder}
            required
          />
        </>
      )}
      {!useDefault && (
        <>
          <Form.Field
            name="aws_iot.endpoint_address"
            type="toggled-input"
            title={m.endpointAddress}
            component={Input.Toggled}
            description={m.endpointAddressDescription}
          />
          <Form.SubTitle title={m.accessKey} />
          <Form.Field
            name="aws_iot._use_access_key"
            title={m.useAccessKey}
            component={Checkbox}
            onChange={handleUseAccessKeyChange}
          />
          {useAccessKey && (
            <>
              <Form.Field
                name="aws_iot.access_key.access_key_id"
                title={m.accessKeyID}
                component={Input}
                required
              />
              <Form.Field
                name="aws_iot.access_key.secret_access_key"
                title={m.accessKeySecret}
                component={Input}
              />
              <Form.Field
                name="aws_iot.access_key.session_token"
                title={m.accessKeySessionToken}
                component={Input}
              />
            </>
          )}
          <Form.SubTitle title={m.assumeRole} />
          <Form.Field
            name="aws_iot._use_assume_role"
            title={m.useAssumeRole}
            component={Checkbox}
            onChange={handleUseAssumeRoleChange}
          />
          {useAssumeRole && (
            <>
              <Form.Field
                name="aws_iot.assume_role.arn"
                title={m.assumeRoleArn}
                component={Input}
                required
              />
              <Form.Field
                name="aws_iot.assume_role.external_id"
                title={m.assumeRoleExternalID}
                component={Input}
              />
              <Form.Field
                name="aws_iot.assume_role.session_duration"
                title={m.assumeRoleSessionDuration}
                component={UnitInput}
                units={[
                  { label: sharedMessages.minutes, value: 'm' },
                  { label: sharedMessages.hours, value: 'h' },
                ]}
              />
            </>
          )}
        </>
      )}
    </>
  )
}

AWSIoTSettings.propTypes = {
  onUseDefaultsChange: PropTypes.func,
  value: PropTypes.pubsubAWSIoT.isRequired,
}

AWSIoTSettings.defaultProps = {
  onUseDefaultsChange: () => null,
}

export default AWSIoTSettings
