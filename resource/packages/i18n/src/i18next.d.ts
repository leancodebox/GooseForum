import type Resources from '../generated/resources'

declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'site.auth'
    parseInterpolation: 'icu'
    resources: Resources
    strictKeyChecks: true
  }
}
