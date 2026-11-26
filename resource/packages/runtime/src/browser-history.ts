// Both layouts share the same session history and its scroll positions.
const positions = new Map<number, { left: number; top: number }>()

export function historyIndex(state: unknown = window.history.state): number {
  if (!state || typeof state !== 'object' || !('gooseEntryIndex' in state)) return 0
  const value = state.gooseEntryIndex
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0 ? value : 0
}

export function commitHistory(url: string | URL, action: 'push' | 'replace', index = historyIndex()) {
  const next = action === 'push' ? index + 1 : index
  const previous: unknown = window.history.state
  const state = { ...(typeof previous === 'object' && previous), gooseEntryIndex: next }
  if (action === 'push') window.history.pushState(state, '', url)
  else window.history.replaceState(state, '', url)
  return next
}

export function rememberScroll(index: number) {
  positions.set(index, { left: window.scrollX, top: window.scrollY })
}

export function restoreScroll(index?: number, hash = '') {
  if (hash) {
    let id = hash.slice(1)
    try { id = decodeURIComponent(id) } catch { /* Literal fragment. */ }
    const anchor = document.getElementById(id)
    if (anchor) { anchor.scrollIntoView({ behavior: 'instant' }); return }
  }
  const saved = index === undefined ? undefined : positions.get(index)
  window.scrollTo({ left: saved?.left ?? 0, top: saved?.top ?? 0, behavior: 'instant' })
}

export function listenToHistory(listener: (event: PopStateEvent) => void) {
  const previous = history.scrollRestoration
  history.scrollRestoration = 'manual'
  window.addEventListener('popstate', listener)
  return () => {
    window.removeEventListener('popstate', listener)
    history.scrollRestoration = previous
  }
}
