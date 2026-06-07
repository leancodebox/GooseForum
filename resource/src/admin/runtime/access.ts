export enum AdminPermission {
  Admin = 0,
  UserManager = 1,
  ArticlesManager = 2,
  PageManager = 3,
  RoleManager = 4,
  SiteManager = 5,
}

const adminPathPermissions: Record<string, AdminPermission> = {
  '/admin': AdminPermission.Admin,
  '/admin/users': AdminPermission.UserManager,
  '/admin/roles': AdminPermission.RoleManager,
  '/admin/categories': AdminPermission.ArticlesManager,
  '/admin/posts': AdminPermission.ArticlesManager,
  '/admin/links': AdminPermission.PageManager,
  '/admin/sponsors': AdminPermission.PageManager,
  '/admin/badges': AdminPermission.SiteManager,
  '/admin/opt-records': AdminPermission.Admin,
  '/admin/settings/site-info': AdminPermission.SiteManager,
  '/admin/settings/mail': AdminPermission.SiteManager,
  '/admin/settings/security': AdminPermission.SiteManager,
  '/admin/settings/posting': AdminPermission.SiteManager,
  '/admin/settings/announcement': AdminPermission.PageManager,
  '/admin/unknown': AdminPermission.PageManager,
}

const adminEntryPaths = [
  '/admin',
  '/admin/users',
  '/admin/posts',
  '/admin/categories',
  '/admin/roles',
  '/admin/links',
  '/admin/sponsors',
  '/admin/badges',
  '/admin/opt-records',
  '/admin/settings/site-info',
  '/admin/settings/mail',
  '/admin/settings/security',
  '/admin/settings/posting',
  '/admin/settings/announcement',
  '/admin/unknown',
]

let permissions = new Set<number>()

export function configureAdminAccess(values: number[] = []) {
  permissions = new Set(values)
}

export function hasAdminPermission(permission: AdminPermission) {
  return permissions.has(AdminPermission.Admin) || permissions.has(permission)
}

export function canVisitAdminPath(path: string) {
  const permission = adminPathPermissions[path]
  return permission !== undefined && hasAdminPermission(permission)
}

export function firstAdminPath() {
  return adminEntryPaths.find(canVisitAdminPath) || '/admin/unknown'
}
