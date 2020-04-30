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

import React, { Component } from 'react'
import PropTypes from 'prop-types'

import style from './video.styl'

class Video extends Component {
  static propTypes = {
    onInit: PropTypes.func,
  }

  static defaultProps = {
    onInit: () => null,
  }

  video = React.createRef()
  stream = null

  async componentDidMount() {
    const video = this.video.current
    const { onInit } = this.props
    const ua = navigator.userAgent.toLowerCase()
    const devices = await navigator.mediaDevices.enumerateDevices()
    const cameras = devices.filter(device => device.kind === 'videoinput')
    const videoMode =
      cameras.length > 1
        ? ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1
          ? { facingMode: { exact: 'environment' } }
          : { deviceId: cameras[1].deviceId }
        : { facingMode: 'environment' }

    this.stream = await navigator.mediaDevices.getUserMedia({
      video: { ...videoMode },
    })
    video.srcObject = this.stream
    requestAnimationFrame(() => {
      onInit(video)
    })
  }

  componentWillUnmount() {
    if (this.stream) this.stream.getTracks().map(t => t.stop())
  }

  render() {
    return (
      <div className={style.container}>
        <video autoPlay playsInline ref={this.video} className={style.video} />
      </div>
    )
  }
}

export default Video
