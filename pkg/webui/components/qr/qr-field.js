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
import QrReader from 'react-qr-reader'
import PropTypes from '../../lib/prop-types'

export default class QrField extends Component {
  state = {
    result: 'No result',
  }

  handleScan = data => {
    const { onChange } = this.props
    if (data) {
      this.setState({ result: data })
      onChange(data)
    }
  }

  handleError = data => {
    console.log(data)
  }
  render() {
    const { result } = this.props
    return (
      <div>
        <QrReader
          delay={200}
          onScan={this.handleScan}
          onError={this.handleError}
          style={{ width: '80%' }}
          showViewFinder={false}
        />
        <p>{result}</p>
      </div>
    )
  }
}

QrField.propTypes = {
  onChange: PropTypes.func.isRequired,
}
