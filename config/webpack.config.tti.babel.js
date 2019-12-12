// Copyright © 2019 The Things Industries B.V.

/* eslint-env node */

import ttnConfig from './webpack.config.babel'

export default {
  ...ttnConfig,
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
