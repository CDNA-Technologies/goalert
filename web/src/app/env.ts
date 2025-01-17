/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable camelcase */

declare let __webpack_public_path__: string
declare let __webpack_require__: any

export const pathPrefix = global.pathPrefix
export const applicationName = global.applicationName

if (typeof __webpack_require__ !== 'undefined')
  // eslint-disable-next-line
  __webpack_require__.p = __webpack_public_path__ = pathPrefix.replace(
    /\/?$/,
    '/',
  )

export const GOALERT_VERSION = global.GOALERT_VERSION

export const isCypress = Boolean(global.Cypress)
