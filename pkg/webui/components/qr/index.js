// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/* eslint-disable no-invalid-this */

import React from 'react'
import PropTypes from 'prop-types'
import jsQR from 'jsqr'
import VideoStream from './video-stream'

import style from './qr.styl'

export default class QR extends React.Component {
  state = {
    result: {
      data: 'No Result',
      location: null,
    },
  }

  onVideoStreamInit = (state, drawFrame) => {
    const { onInit } = this.props

    if (onInit) {
      onInit(state)
    }
    this.drawVideoFrame = drawFrame
    this.drawVideoFrame()
  }

  onFrame = event => {
    const { data, width, height } = event
    const code = jsQR(data, width, height)
    this.onFrameDecoded(code)
  }

  onFrameDecoded = code => {
    const { onChange } = this.props
    const { result } = this.state

    if (code !== null) {
      const { data } = code
      if (data.length > 0) {
        this.setState({ result: code })
        onChange(data)
      }
    } else {
      this.setState({
        result: {
          ...result,
          location: null,
        },
      })
    }

    this.drawVideoFrame()
  }

  render() {
    const { result } = this.state
    return (
      <div className={style.container}>
        <VideoStream
          onFrame={this.onFrame}
          onInit={this.onVideoStreamInit}
          location={result.location}
        />
        <p className={style.result}>{result.data}</p>
      </div>
    )
  }
}

QR.propTypes = {
  onChange: PropTypes.func.isRequired,
  onInit: PropTypes.func,
}

QR.defaultProps = {
  onInit: () => null,
}
