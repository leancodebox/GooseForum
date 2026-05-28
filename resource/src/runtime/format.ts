import { i18n } from './i18n'

export function formatNumber(value: number): string {
  if (value >= 1000000) return `${trim(value / 1000000)}m`
  if (value >= 1000) return `${trim(value / 1000)}k`
  return String(value)
}

export function formatDateTime(value: string): string {
  if (!value) return ''
  const date = parseDate(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatChatTime(value: string): string {
  if (!value) return ''
  const date = parseDate(value)
  if (Number.isNaN(date.getTime())) return value
  const now = new Date()
  const time = `${pad(date.getHours())}:${pad(date.getMinutes())}`
  if (isSameDay(date, now)) return time
  if (date.getFullYear() === now.getFullYear()) {
    return i18n.global.t('date.monthDayTime', { month: date.getMonth() + 1, day: date.getDate(), time })
  }
  return i18n.global.t('date.yearMonthDayTime', { year: date.getFullYear(), month: date.getMonth() + 1, day: date.getDate(), time })
}

export function timeAgo(value: string): string {
  if (!value) return ''
  const timestamp = parseDate(value).getTime()
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
  const date = parseDate(value)
  if (Number.isNaN(date.getTime())) return value.split(' ')[0] || value
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function parseDate(value: string): Date {
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  return new Date(normalized)
}

function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear()
    && a.getMonth() === b.getMonth()
    && a.getDate() === b.getDate()
}

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

function trim(value: number): string {
  return value.toFixed(value >= 10 ? 0 : 1).replace(/\.0$/, '')
}
