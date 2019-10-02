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

export default class Video extends React.Component {
  constructor(props) {
    super(props)
    this.video = React.createRef()
    this.stream = null
    this.ctx = null
  }
  async componentWillMount() {
    const { videoMode, width, height, onInit } = this.props
    try {
      this.stream = await navigator.mediaDevices.getUserMedia({
        audio: false,
        video: videoMode, // Needs to be extended to support back camera and other browsers, but this can be passed as a prop
      })

      if (this.video.current.srcObject !== undefined) {
        this.video.current.playsInline = true
        this.video.current.srcObject = this.stream
        this.video.current.setAttribute('width', width)
        this.video.current.setAttribute('height', height)
      }

      if (!this.canvasContext) {
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        this.canvasContext = canvas.getContext('2d')
      }
    } catch (e) {}

    if (typeof onInit === 'function') {
      onInit(this.drawFrame)
    }
  }

  componentWillUnmount() {
    if (this.stream) {
      this.stream.getTracks().map(t => t.stop())
      this.stream = null
    }
  }

  captureFrame = (width, height) => {
    this.canvasContext.drawImage(this.video, 0, 0, width, height)
    return this.canvasContext.getImageData(0, 0, width, height)
  }

  drawFrame = () => {
    window.requestAnimationFrame(() => {
      if (!this.canvasContext) return
      const { data } = this.captureFrame()
      const { width, height } = this.props()
      this.props.onFrame({
        data,
        width,
        height,
      })
    })
  }

  render() {
    return <video ref={this.video} autoPlay />
  }
}

Video.propTypes = {
  height: PropTypes.number,
  onFrame: PropTypes.func.isRequired,
  onInit: PropTypes.func.isRequired,
  videoMode: PropTypes.shape({}),
  width: PropTypes.number,
}

Video.defaultProps = {
  videoMode: { facingMode: 'user' },
  width: 500,
  height: 500,
}
