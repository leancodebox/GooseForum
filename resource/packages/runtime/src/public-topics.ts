import { useSyncExternalStore } from 'react'

let latest: number | undefined
const listeners = new Set<() => void>()
export function announcePublicTopicMax(id: unknown) {
  if (typeof id !== 'number' || !Number.isSafeInteger(id) || id < 0 || id === latest) return
  latest = id
  listeners.forEach(listener => listener())
}
const subscribe = (listener: () => void) => { listeners.add(listener); return () => { listeners.delete(listener) } }
export function usePublicTopicMax() {
  return useSyncExternalStore(subscribe, () => latest, () => undefined)
}
