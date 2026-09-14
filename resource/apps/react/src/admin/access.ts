export const AdminPermission = {
  Admin: 0,
  UserManager: 1,
  TopicsManager: 2,
  PageManager: 3,
  RoleManager: 4,
  SiteManager: 5,
} as const

export type AdminPermission = typeof AdminPermission[keyof typeof AdminPermission]

export function hasAdminPermission(values: number[], required: AdminPermission) {
  return values.includes(AdminPermission.Admin) || values.includes(required)
}

export function hasAnyAdminPermission(values: number[], required: AdminPermission | AdminPermission[]) {
  return (Array.isArray(required) ? required : [required]).some((item) => hasAdminPermission(values, item))
}

const pathPermissions: Record<string, AdminPermission | AdminPermission[]> = {
  '/admin': AdminPermission.Admin,
  '/admin/users': AdminPermission.UserManager,
  '/admin/roles': AdminPermission.RoleManager,
  '/admin/access-groups': AdminPermission.RoleManager,
  '/admin/categories': [AdminPermission.TopicsManager, AdminPermission.RoleManager],
  '/admin/posts': AdminPermission.TopicsManager,
  '/admin/links': AdminPermission.PageManager,
  '/admin/sponsors': AdminPermission.PageManager,
  '/admin/badges': AdminPermission.SiteManager,
  '/admin/files/resources': AdminPermission.SiteManager,
  '/admin/opt-records': AdminPermission.Admin,
  '/admin/settings/site-info': AdminPermission.SiteManager,
  '/admin/settings/site-chrome': AdminPermission.SiteManager,
  '/admin/settings/mail': AdminPermission.SiteManager,
  '/admin/settings/oauth': AdminPermission.SiteManager,
  '/admin/settings/oidc-provider': AdminPermission.SiteManager,
  '/admin/settings/security': AdminPermission.SiteManager,
  '/admin/settings/posting': AdminPermission.SiteManager,
  '/admin/settings/announcement': AdminPermission.PageManager,
  '/admin/settings/http-notify': AdminPermission.SiteManager,
  '/admin/settings/sensitive-words': AdminPermission.SiteManager,
}

export function canVisitAdminPath(values: number[], path: string) {
  const required = pathPermissions[path]
  return required !== undefined && hasAnyAdminPermission(values, required)
}

export function firstAdminPath(values: number[]) {
  return Object.keys(pathPermissions).find((path) => canVisitAdminPath(values, path)) || '/'
}
