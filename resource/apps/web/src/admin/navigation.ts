export function normalizeAdminPath(pathname: string) {
  return pathname.replace(/\/+$/, '') || '/admin'
}

export const communityRulePaths = ['/admin/settings/security', '/admin/settings/posting', '/admin/settings/trust'] as const
export const identitySettingPaths = ['/admin/settings/oauth', '/admin/settings/oidc-provider'] as const

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
