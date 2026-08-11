import { afterEach, describe, expect, it } from 'vitest'
import { AdminPermission, canVisitAdminPath, configureAdminAccess, hasAnyAdminPermission } from '@/admin/runtime/access'

afterEach(() => configureAdminAccess([]))

describe('admin route permissions', () => {
  it('allows role managers to configure category access without topic management', () => {
    configureAdminAccess([AdminPermission.RoleManager])

    expect(canVisitAdminPath('/admin/categories')).toBe(true)
    expect(canVisitAdminPath('/admin/access-groups')).toBe(true)
    expect(canVisitAdminPath('/admin/posts')).toBe(false)
  })

  it('keeps category maintenance available to topic managers without role management', () => {
    configureAdminAccess([AdminPermission.TopicsManager])

    expect(canVisitAdminPath('/admin/categories')).toBe(true)
    expect(canVisitAdminPath('/admin/access-groups')).toBe(false)
    expect(canVisitAdminPath('/admin/posts')).toBe(true)
  })

  it('treats admin as satisfying any alternative permission', () => {
    configureAdminAccess([AdminPermission.Admin])

    expect(hasAnyAdminPermission([AdminPermission.TopicsManager, AdminPermission.RoleManager])).toBe(true)
    expect(canVisitAdminPath('/admin/categories')).toBe(true)
  })
})
