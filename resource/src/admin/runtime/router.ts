import { ref } from 'vue'

function normalizePath(path: string): string {
  const normalized = path.replace(/\/+$/, '') || '/admin'
  return normalized.startsWith('/admin') ? normalized : '/admin'
}

export const currentAdminPath = ref(normalizePath(window.location.pathname))

let listening = false

export function installAdminRouter() {
  if (listening) return
  listening = true
  window.addEventListener('popstate', () => {
    currentAdminPath.value = normalizePath(window.location.pathname)
  })
}

export function navigateAdmin(path: string, event?: MouseEvent) {
  if (event) {
    event.preventDefault()
  }
  const nextPath = normalizePath(path)
  if (nextPath === currentAdminPath.value) return
  window.history.pushState({}, '', nextPath)
  currentAdminPath.value = nextPath
}
