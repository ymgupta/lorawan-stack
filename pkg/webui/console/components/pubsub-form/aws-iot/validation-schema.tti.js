// Copyright © 2020 The Things Industries B.V.

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from './messages.tti'
import regions from './regions.tti'

const endpointAddressRegexp = /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])$/
const accessKeyIDRegexp = /^[\w]*$/
const assumeRoleArnRegexp = /^arn:aws:iam::[0-9]{12}:role\/[A-Za-z0-9_+=,.@-]+$/
const externalIDRegexp = /^[\w+=,.@:/-]*$/
const stackNameRegexp = /^[A-Za-z][A-Za-z0-9-]*$/
const sessionDurationRegexp = /^[0-9]{1,}[.]?([0-9]{1,})?[a-zA-Z]{1,2}$/

export default Yup.object().shape({
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
