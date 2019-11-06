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
import VideoStream from './video-stream'
// eslint-disable-next-line import/default
import Worker from './qr_decode.worker'

import style from './qr.styl'

export default class QR extends React.Component {
  state = {
    result: {
      data: 'No Result',
      location: null,
    },
  }

  webWorker = null
  location = null

  componentWillMount() {
    this.webWorker = new Worker()
    this.webWorker.addEventListener('message', this.onFrameDecoded)
  }

  componentWillUnmount() {
    if (this.webWorker !== null) {
      this.webWorker.terminate()
      this.webWorker = null
    }
  }

  onVideoStreamInit = (state, drawFrame) => {
    const { onInit } = this.props

    if (onInit) {
      onInit(state)
    }
    this.drawVideoFrame = drawFrame
    this.drawVideoFrame()
  }

  onFrame = frameData => this.webWorker.postMessage(frameData)
  drawVideoFrame = () => {}

  onFrameDecoded = event => {
    const { onChange } = this.props
    const { result } = this.state
    const code = event.data

    if (code && code.binaryData) {
      const { data } = code
      if (data.length > 0) {
        this.setState({ result: code })
        onChange(code.data)
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
          rearCamera={this.props.rearCamera}
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
  rearCamera: PropTypes.bool,
}

QR.defaultProps = {
  onInit: () => {},
  rearCamera: true,
}
