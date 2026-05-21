import { readonly, ref } from 'vue'

export type FlashMessageType = 'success' | 'info' | 'warning' | 'error'

export interface FlashMessage {
  id: number
  type: FlashMessageType
  message: string
}

const STORAGE_KEY = 'goose:flash-messages'
const messages = ref<FlashMessage[]>([])
let nextId = 1
let hydrated = false

function readStoredMessages(): Omit<FlashMessage, 'id'>[] {
  try {
    const raw = window.sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    window.sessionStorage.removeItem(STORAGE_KEY)
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .map((item) => ({
        type: normalizeType(item?.type),
        message: String(item?.message || '').trim(),
      }))
      .filter((item) => item.message)
  } catch {
    return []
  }
}

function normalizeType(type: unknown): FlashMessageType {
  if (type === 'success' || type === 'info' || type === 'warning' || type === 'error') {
    return type
  }
  return 'info'
}

function push(message: string, type: FlashMessageType = 'info') {
  const text = message.trim()
  if (!text) return
  const item: FlashMessage = {
    id: nextId++,
    type,
    message: text,
  }
  messages.value = [...messages.value, item]
  window.setTimeout(() => dismiss(item.id), 5200)
}

export function queueFlashMessage(message: string, type: FlashMessageType = 'info') {
  const text = message.trim()
  if (!text) return
  try {
    const raw = window.sessionStorage.getItem(STORAGE_KEY)
    const existing = raw ? JSON.parse(raw) : []
    const list = Array.isArray(existing) ? existing : []
    list.push({ type, message: text })
    window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(list.slice(-4)))
  } catch {
    push(text, type)
  }
}

export function hydrateFlashMessages() {
  if (hydrated) return
  hydrated = true
  readStoredMessages().forEach((item) => push(item.message, item.type))
}

export function dismiss(id: number) {
  messages.value = messages.value.filter((item) => item.id !== id)
}

export function useFlashMessages() {
  return {
    messages: readonly(messages),
    push,
    dismiss,
  }
}
