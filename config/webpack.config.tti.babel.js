// Copyright © 2019 The Things Industries B.V.

/* eslint-env node */

import path from 'path'
import { merge } from 'lodash'

import ttnConfig from './webpack.config.babel'

const { CONTEXT = '.' } = process.env

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

config.devServer.proxy.push({
  context: ['/claim'],
  target: WEBPACK_DEV_SERVER_USE_TLS ? 'https://localhost:8885' : 'http://localhost:1885',
  changeOrigin: true,
})

export default config
