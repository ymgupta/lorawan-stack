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
import bind from 'autobind-decorator'
import * as jsQR from 'jsqr'
import Video from './input/video'
import Capture from './input/capture'

export default class QR extends React.Component {
  static propTypes = {
    onChange: PropTypes.func.isRequired,
  }

  canvas = document.createElement('canvas')
  ctx = this.canvas.getContext('2d')
  state = {
    value: null,
  }

  @bind
  handleVideoInit(video) {
    const { videoWidth: width, videoHeight: height } = video
    this.handleRead(video, width, height)
    requestAnimationFrame(() => {
      // Loop over video feed.
      this.handleVideoInit(video)
    })
  }

  @bind
  handleRead(media, width, height) {
    const { onChange } = this.props
    const { value } = this.state

    if (!width && !height) return

    this.canvas.width = width
    this.canvas.height = height
    this.ctx.drawImage(media, 0, 0, width, height)

    const { data } = this.ctx.getImageData(0, 0, width, height)
    const qr = jsQR(data, width, height, {
      // !Important dontInvert fixes a ~50% performance hit.
      inversionAttempts: 'dontInvert',
    })

    if (qr && qr.data !== value) {
      this.setState({ value: qr.data })
      onChange(qr.data, true)
    }
  }

  render() {
    return !!navigator.mediaDevices ? (
      <Video onInit={this.handleVideoInit} />
    ) : (
      <Capture onRead={this.handleRead} />
    )
  }
}
