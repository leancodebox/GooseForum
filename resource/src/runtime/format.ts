import { i18n } from './i18n'

export function formatNumber(value: number): string {
  if (value >= 1000000) return `${trim(value / 1000000)}m`
  if (value >= 1000) return `${trim(value / 1000)}k`
  return String(value)
}

export function formatDateTime(value: string): string {
  if (!value) return ''
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function timeAgo(value: string): string {
  if (!value) return ''
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const timestamp = new Date(normalized).getTime()
  if (Number.isNaN(timestamp)) return value
  const seconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
  if (seconds < 60) return i18n.global.t('time.justNow')
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return i18n.global.t('time.minuteAgo', { count: minutes })
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return i18n.global.t('time.hourAgo', { count: hours })
  const days = Math.floor(hours / 24)
  if (days < 7) return i18n.global.t('time.dayAgo', { count: days })
  return formatDate(value)
}

export function formatDate(value: string): string {
  if (!value) return ''
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) return value.split(' ')[0] || value
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function trim(value: number): string {
  return value.toFixed(value >= 10 ? 0 : 1).replace(/\.0$/, '')
}
