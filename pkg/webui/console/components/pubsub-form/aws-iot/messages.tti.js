// Copyright © 2020 The Things Industries B.V.

import { defineMessages } from 'react-intl'

export default defineMessages({
  config: 'AWS IoT configuration',
  useDefault: 'Use default integration',
  defaultInfo:
    'The default AWS IoT integration can be deployed via CloudFormation in your AWS account. {moreInformation}',
  region: 'Region',
  endpointAddress: 'Endpoint address',
  endpointAddressDescription:
    'If the endpoint address is left empty, the integration will try to discover it',
  accessKey: 'Access key',
  accessKeyID: 'Access key ID',
  accessKeySecret: 'Secret access key',
  accessKeySessionToken: 'Session token',
  assumeRole: 'Assume role',
  assumeRoleArn: 'Role ARN',
  assumeRoleArnPlaceholder: 'arn:aws:iam::123456789012:role/...',
  assumeRoleExternalID: 'External ID',
  assumeRoleSessionDuration: 'Session duration',
  defaultStackName: 'CloudFormation stack name',
  defaultStackNameDescription:
    'Copy the CloudFormation stack name that you used when deploying the integration in your AWS account',
  defaultRoleArn: 'Cross-account role ARN',
  defaultRoleArnDescription: 'Copy the CloudFormation stack "CrossAccountRoleArn" output',
  validateAccessKeyIDFormat: '{field} must be a valid AWS access key ID',
  validateEndpointAddressFormat: '{field} must be a valid AWS IoT Core endpoint address',
  validateRoleArnFormat: '{field} must be a valid AWS IAM role ARN',
  validateExternalIDFormat: '{field} must be a valid AWS IAM role external ID',
  validateSessionDurationFormat: '{field} must be a valid AWS IAM session duration',
  validateStackNameFormat: '{field} must be a valid AWS CloudFormation stack name',
})
