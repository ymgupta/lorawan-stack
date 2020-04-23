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

import React from 'react'
import classnames from 'classnames'

import Link from '@ttn-lw/components/link'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './logo.styl'

const Logo = function({ className, logo, secondaryLogo, vertical, text, clusterTag }) {
  const classname = classnames(style.container, className, {
    [style.vertical]: vertical,
    [style.customBranding]: Boolean(secondaryLogo),
  })
  const cappedText = text ? text.substr(0, 25) : undefined
  const cappedClusterTag = clusterTag ? clusterTag.substr(0, 5) : undefined

  return (
    <div className={classname}>
      <div className={style.logo}>
        <Link className={style.logoContainer} to="/">
          <img {...logo} />
        </Link>
      </div>
      {Boolean(secondaryLogo) && (
        <div className={style.secondaryLogo}>
          <div className={style.secondaryLogoContainer}>
            <img {...secondaryLogo} />
          </div>
          {text && (
            <div className={style.text}>
              {clusterTag && <span className={style.clusterTag}>{cappedClusterTag}</span>}
              {cappedText}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Logo.propTypes = {
  className: PropTypes.string,
  clusterTag: PropTypes.string,
  logo: imgPropType.isRequired,
  secondaryLogo: imgPropType,
  text: PropTypes.string,
  vertical: PropTypes.bool,
}

Logo.defaultProps = {
  className: undefined,
  clusterTag: undefined,
  secondaryLogo: undefined,
  text: undefined,
  vertical: false,
}

export default Logo
