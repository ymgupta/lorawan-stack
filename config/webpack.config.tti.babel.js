// Copyright © 2019 The Things Industries B.V.

/* eslint-env node */

import path from 'path'

import ttnConfig from './webpack.config.babel'

const { CONTEXT = '.' } = process.env

const context = path.resolve(CONTEXT)

export default {
  ...ttnConfig,
  resolve: {
    alias: {
      ...ttnConfig.resolve.alias,
      '@claim': path.resolve(context, 'pkg/webui/claim'),
    },
  },
  entry: {
    ...ttnConfig.entry,
    claim: ['./config/root.js', './pkg/webui/claim.js'],
  },
  output: {
    ...ttnConfig.output,
    globalObject: 'this',
  },
  devServer: {
    ...ttnConfig.devServer,
    proxy: [
      ...ttnConfig.devServer.proxy,
      {
        context: ['/claim'],
        target: 'http://localhost:1885',
        changeOrigin: true,
      },
    ],
  },
}
