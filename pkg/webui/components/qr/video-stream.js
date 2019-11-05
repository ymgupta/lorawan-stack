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

class VideoStream extends Component {
  stream = null
  streamWidth = 0
  streamHeight = 0
  video = React.createRef()
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
    let videoMode = { facingMode: 'user' }
    if (cameras.length > 1) {
      const cameraIndex = this.props.rearCamera ? 1 : 0
      const cameraEnv = this.props.rearCamera ? 'environment' : 'user'
      videoMode =
        ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1
          ? { facingMode: { exact: cameraEnv } }
          : { deviceId: cameras[cameraIndex].deviceId }
    }

    this.stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: videoMode,
    })

    if (this.video.current.srcObject !== undefined) {
      this.video.current.srcObject = this.stream
    } else if (this.video.current.mozSrcObject !== undefined) {
      this.video.current.mozSrcObject = this.stream
    } else if (window.URL.createObjectURL) {
      this.video.current.src = window.URL.createObjectURL(this.stream)
    } else if (window.webkitURL) {
      this.video.current.src = window.webkitURL.createObjectURL(this.stream)
    } else {
      this.video.current.src = this.stream
    }

    this.video.current.playsInline = true
    this.video.current.play() // firefox does not emit `loadeddata` if video not playing
    await this.streamLoadedPromise()

    this.streamWidth = this.video.current.videoWidth
    this.streamHeight = this.video.current.videoHeight

    if (!this.canvasContext) {
      const canvas = document.createElement('canvas')
      canvas.width = this.streamWidth
      canvas.height = this.streamHeight
      this.canvasContext = canvas.getContext('2d')
    }
  }

  streamLoadedPromise = () =>
    new Promise((resolve, reject) => {
      this.video.current.addEventListener('loadeddata', resolve, { once: true })
      this.video.current.addEventListener('error', reject, { once: true })
    })

  captureFrame = () => {
    this.canvasContext.drawImage(this.video.current, 0, 0, this.streamWidth, this.streamHeight)
    return this.canvasContext.getImageData(0, 0, this.streamWidth, this.streamHeight)
  }

  drawFrame = () => {
    window.requestAnimationFrame(() => {
      if (!this.canvasContext) return
      const { data } = this.captureFrame()
      this.props.onFrame({
        data,
        width: this.streamWidth,
        height: this.streamHeight,
      })
    })
  }

  render() {
    return <video ref={this.video} className={style.video} autoPlay />
  }
}

VideoStream.propTypes = {
  onFrame: PropTypes.func.isRequired,
  onInit: PropTypes.func.isRequired,
  rearCamera: PropTypes.bool,
}

VideoStream.defaultProps = {
  rearCamera: true,
}

export default VideoStream
