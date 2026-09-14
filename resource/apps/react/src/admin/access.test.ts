import { describe, expect, it } from 'vitest'
import {
  AdminPermission,
  canVisitAdminPath,
  firstAdminPath,
  hasAdminPermission,
} from './access'
import { normalizeAdminPath, shouldUseClientNavigation } from './navigation'

describe('React admin access', () => {
  it('treats the administrator permission as a superuser permission', () => {
    expect(hasAdminPermission([AdminPermission.Admin], AdminPermission.SiteManager)).toBe(true)
  })

  it('allows either topic or role managers to open categories', () => {
    expect(canVisitAdminPath([AdminPermission.TopicsManager], '/admin/categories')).toBe(true)
    expect(canVisitAdminPath([AdminPermission.RoleManager], '/admin/categories')).toBe(true)
    expect(canVisitAdminPath([AdminPermission.PageManager], '/admin/categories')).toBe(false)
  })

  it('selects the first route available to a limited administrator', () => {
    expect(firstAdminPath([AdminPermission.RoleManager])).toBe('/admin/roles')
    expect(firstAdminPath([AdminPermission.TopicsManager])).toBe('/admin/categories')
  })

  it('normalizes admin URLs without breaking the admin root', () => {
    expect(normalizeAdminPath('/admin/categories/')).toBe('/admin/categories')
    expect(normalizeAdminPath('/admin/')).toBe('/admin')
  })

  it('keeps modified and non-primary clicks as regular browser navigation', () => {
    const click = { button: 0, defaultPrevented: false, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false }
    expect(shouldUseClientNavigation(click)).toBe(true)
    expect(shouldUseClientNavigation({ ...click, metaKey: true })).toBe(false)
    expect(shouldUseClientNavigation({ ...click, button: 1 })).toBe(false)
  })
})
