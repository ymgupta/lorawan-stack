// Copyright © 2020 The Things Industries B.V.

import { merge } from 'lodash'

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
