// Copyright © 2019 The Things Industries B.V.

/* eslint-env node */

import path from 'path'
import { merge } from 'lodash'

import ttnConfig from './webpack.config.babel'

const { CONTEXT = '.' } = process.env

const WEBPACK_DEV_SERVER_ENABLE_MULTI_TENANCY =
  process.env.WEBPACK_DEV_SERVER_ENABLE_MULTI_TENANCY === 'true'
const WEBPACK_DEV_SERVER_USE_TLS = process.env.WEBPACK_DEV_SERVER_USE_TLS === 'true'

const context = path.resolve(CONTEXT)

const config = merge(ttnConfig, {
  resolve: {
    alias: {
      '@claim': path.resolve(context, 'pkg/webui/claim'),
    },
  },
  entry: {
    claim: ['./config/root.js', './pkg/webui/claim.js'],
  },
  output: {
    globalObject: 'this',
  },
})

// Add claim app to proxies.
config.devServer.proxy[0].context.push('/claim')

if (WEBPACK_DEV_SERVER_ENABLE_MULTI_TENANCY) {
  // Use default proxy only for api route.
  config.devServer.proxy[0].context = '/api'

  // Add new proxy using default tenant URL.
  config.devServer.disableHostCheck = true
  config.devServer.headers = { 'Access-Control-Allow-Origin': '*' }
  config.devServer.proxy.push({
    context: ['/console', '/oauth', '/claim'],
    target: WEBPACK_DEV_SERVER_USE_TLS
      ? 'https://default.localhost:8885'
      : 'http://default.localhost:1885',
    changeOrigin: true,
    secure: false,
  })
}

export default config
