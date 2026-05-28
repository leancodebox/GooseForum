import { i18n } from './i18n'

export interface ApiMessageEnvelope {
  messageCode?: string
  params?: Record<string, unknown>
}

function fallbackText(value: unknown) {
  return typeof value === 'string' && value.trim() ? value : undefined
}

function interpolate(template: string, params: Record<string, unknown> = {}) {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => {
    const value = params[key]
    return value === undefined || value === null ? '' : String(value)
  })
}

function flatServerMessage(messageCode: string, params: Record<string, unknown>) {
  const locale = i18n.global.locale.value
  const messages = i18n.global.getLocaleMessage(locale) as { serverMessages?: Record<string, string> }
  const template = messages.serverMessages?.[messageCode]
  return template ? interpolate(template, params) : undefined
}

function localizeParams(params: Record<string, unknown>) {
  const actionCode = fallbackText(params.actionCode)
  if (!actionCode) return params

  const actionKey = `serverActions.${actionCode}`
  if (!i18n.global.te(actionKey)) return params

  return {
    ...params,
    action: i18n.global.t(actionKey),
  }
}

export function resolveApiMessage(data: ApiMessageEnvelope, fallback: string) {
  const messageCode = fallbackText(data.messageCode)
  if (messageCode) {
    const params = localizeParams(data.params || {})
    const flatMessage = flatServerMessage(messageCode, params)
    if (flatMessage) return flatMessage

    const key = `server.${messageCode}`
    if (i18n.global.te(key)) {
      return i18n.global.t(key, params)
    }
  }
  return fallback
}
