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

import React, { Component } from 'react'
import PropTypes from 'prop-types'

import style from './video.styl'

class Video extends Component {
  stream = null
  streamWidth = 0
  streamHeight = 0
  video = document.createElement('video')
  canvas = React.createRef()
  canvasContext = null

  async componentDidMount() {
    let initSuccess = true
    let message = ''
    try {
      await this.startCamera()
    } catch (e) {
      message = `Browser camera init error: ${e}`
      initSuccess = false
    }

    if (typeof this.props.onInit === 'function') {
      this.props.onInit({ error: initSuccess, message }, this.drawFrame)
    }
  }

  componentWillUnmount() {
    this.stopCamera()
  }

  stopCamera = () => {
    if (!this.stream) return
    this.stream.getTracks().map(t => t.stop())
    this.stream = null
    this.streamWidth = 0
    this.streamHeight = 0
    this.canvasContext = null
  }

  startCamera = async () => {
    this.stopCamera()
    const ua = navigator.userAgent.toLowerCase()

    if (!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia)) {
      throw new Error('WebRTC API not supported in this browser')
    }

    const devices = await navigator.mediaDevices.enumerateDevices()
    const cameras = devices.filter(device => device.kind === 'videoinput')
    let videoMode = { facingMode: 'environment' }
    if (cameras.length > 1) {
      videoMode =
        ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1
          ? { facingMode: { exact: 'environment' } }
          : { deviceId: cameras[1].deviceId }
    }

    this.stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: {
        ...videoMode,
        frameRate: { ideal: 10, max: 15 },
        width: { ideal: 1280 },
        height: { ideal: 720 },
      },
    })

    if (this.video.srcObject !== undefined) {
      this.video.srcObject = this.stream
    } else if (this.video.mozSrcObject !== undefined) {
      this.video.mozSrcObject = this.stream
    } else if (window.URL.createObjectURL) {
      this.video.src = window.URL.createObjectURL(this.stream)
    } else if (window.webkitURL) {
      this.video.src = window.webkitURL.createObjectURL(this.stream)
    } else {
      this.video.src = this.stream
    }

    this.video.playsInline = true
    this.video.play() // firefox does not emit `loadeddata` if video not playing
    await this.streamLoadedPromise()

    this.streamWidth = this.video.videoWidth
    this.streamHeight = this.video.videoHeight

    if (!this.canvasContext) {
      this.canvas.current.width = this.streamWidth
      this.canvas.current.height = this.streamHeight
      this.canvasContext = this.canvas.current.getContext('2d')
    }
  }

  drawLine = (begin, end, color) => {
    this.canvasContext.beginPath()
    this.canvasContext.moveTo(begin.x, begin.y)
    const midX = begin.x + (end.x - begin.x) * 0.25
    const midY = begin.y + (end.y - begin.y) * 0.25

    this.canvasContext.lineTo(midX, midY)
    this.canvasContext.lineWidth = 4
    this.canvasContext.strokeStyle = color
    this.canvasContext.stroke()
  }

  streamLoadedPromise = () =>
    new Promise((resolve, reject) => {
      this.video.addEventListener('loadeddata', resolve, { once: true })
      this.video.addEventListener('error', reject, { once: true })
    })

  captureFrame = () => {
    this.canvasContext.drawImage(this.video, 0, 0, this.streamWidth, this.streamHeight)
    return this.canvasContext.getImageData(0, 0, this.streamWidth, this.streamHeight)
  }

  drawFrame = () => {
    window.requestAnimationFrame(() => {
      this.tick()
    })
  }

  tick = () => {
    const { location, onFrame } = this.props

    if (!this.canvasContext) return

    const { data } = this.captureFrame()

    if (location) {
      this.drawLine(location.topLeftCorner, location.topRightCorner, '#0030b5')
      this.drawLine(location.topLeftCorner, location.bottomLeftCorner, '#0030b5')

      this.drawLine(location.topRightCorner, location.bottomRightCorner, '#0030b5')
      this.drawLine(location.topRightCorner, location.topLeftCorner, '#0030b5')

      this.drawLine(location.bottomRightCorner, location.bottomLeftCorner, '#0030b5')
      this.drawLine(location.bottomRightCorner, location.topRightCorner, '#0030b5')

      this.drawLine(location.bottomLeftCorner, location.topLeftCorner, '#0030b5')
      this.drawLine(location.bottomLeftCorner, location.bottomRightCorner, '#0030b5')
    }

    onFrame({
      data,
      width: this.streamWidth,
      height: this.streamHeight,
    })
  }

  render() {
    return <canvas ref={this.canvas} className={style.video} />
  }
}

Video.propTypes = {
  location: PropTypes.shape({
    bottomLeftCorner: PropTypes.objectOf(PropTypes.number),
    bottomLeftFinderPattern: PropTypes.objectOf(PropTypes.number),
    bottomRightAlignmentPattern: PropTypes.objectOf(PropTypes.number),
    bottomRightCorner: PropTypes.objectOf(PropTypes.number),
    topLeftCorner: PropTypes.objectOf(PropTypes.number),
    topLeftFinderPattern: PropTypes.objectOf(PropTypes.number),
    topRightCorner: PropTypes.objectOf(PropTypes.number),
    topRightFinderPattern: PropTypes.objectOf(PropTypes.number),
  }),
  onFrame: PropTypes.func.isRequired,
  onInit: PropTypes.func.isRequired,
}

Video.defaultProps = {
  location: null,
}

export default Video
