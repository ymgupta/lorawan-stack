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
import FileInput from '../../../file-input'

export default class Capture extends Component {
  // Convert base64 to Uint8ClampedArray
  handleChange = data => {
    const { onFrame } = this.props
    const image = new Image()
    image.src = data
    image.onload = () => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      canvas.width = image.width
      canvas.height = image.height
      ctx.drawImage(image, 0, 0)
      const imageData = ctx.getImageData(0, 0, image.width, image.height)
      onFrame({
        data: imageData.data,
        width: imageData.width,
        height: imageData.height,
      })
    }
  }

  // Do not transform content.
  handleDataTransform = content => content

  render() {
    return (
      <FileInput onChange={this.handleChange} dataTransform={this.handleDataTransform} capture />
    )
  }
}

Capture.propTypes = {
  onFrame: PropTypes.func.isRequired,
}

Capture.defaultProps = {}
