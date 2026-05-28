import { i18n } from '@/runtime/i18n'

export function adminText(key: string, named?: Record<string, unknown>) {
  return i18n.global.t(`adminRaw.${key}`, named || {})
}
