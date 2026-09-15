export function normalizeAdminPath(pathname: string) {
  return pathname.replace(/\/+$/, '') || '/admin'
}

export interface AdminNavigationClick {
  button: number
  defaultPrevented: boolean
  metaKey: boolean
  ctrlKey: boolean
  shiftKey: boolean
  altKey: boolean
}

export function shouldUseClientNavigation(event: AdminNavigationClick) {
  return event.button === 0 &&
    !event.defaultPrevented &&
    !event.metaKey &&
    !event.ctrlKey &&
    !event.shiftKey &&
    !event.altKey
}
