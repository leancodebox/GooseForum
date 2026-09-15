import type { ReactNode } from 'react'
import { Award, FileText, Files, Heart, KeyRound, Link, ListChecks, Mail, Megaphone, Monitor, PanelLeft, PanelsTopLeft, ShieldAlert, ShieldCheck, Tags, UserCog, UsersRound, Webhook } from 'lucide-react'
import { AdminPermission } from './access'
import type { AdminTextKey } from './i18n'

export interface AdminNavItem {
  label: AdminTextKey
  url: string
  icon: ReactNode
  permission: AdminPermission | AdminPermission[]
}

export const adminNavGroups: { label?: AdminTextKey; items: AdminNavItem[] }[] = [
  { items: [
    { label: 'dashboard', url: '/admin', icon: <Monitor />, permission: AdminPermission.Admin },
    { label: 'users', url: '/admin/users', icon: <UserCog />, permission: AdminPermission.UserManager },
    { label: 'roles', url: '/admin/roles', icon: <ShieldCheck />, permission: AdminPermission.RoleManager },
    { label: 'accessGroups', url: '/admin/access-groups', icon: <UsersRound />, permission: AdminPermission.RoleManager },
    { label: 'categories', url: '/admin/categories', icon: <Tags />, permission: [AdminPermission.TopicsManager, AdminPermission.RoleManager] },
    { label: 'posts', url: '/admin/posts', icon: <FileText />, permission: AdminPermission.TopicsManager },
    { label: 'links', url: '/admin/links', icon: <Link />, permission: AdminPermission.PageManager },
    { label: 'sponsors', url: '/admin/sponsors', icon: <Heart />, permission: AdminPermission.PageManager },
    { label: 'badges', url: '/admin/badges', icon: <Award />, permission: AdminPermission.SiteManager },
    { label: 'resources', url: '/admin/files/resources', icon: <Files />, permission: AdminPermission.SiteManager },
    { label: 'records', url: '/admin/opt-records', icon: <ListChecks />, permission: AdminPermission.Admin },
  ] },
  { label: 'settings', items: [
    { label: 'siteInfo', url: '/admin/settings/site-info', icon: <PanelsTopLeft />, permission: AdminPermission.SiteManager },
    { label: 'siteChrome', url: '/admin/settings/site-chrome', icon: <PanelLeft />, permission: AdminPermission.SiteManager },
    { label: 'mail', url: '/admin/settings/mail', icon: <Mail />, permission: AdminPermission.SiteManager },
    { label: 'oauth', url: '/admin/settings/oauth', icon: <KeyRound />, permission: AdminPermission.SiteManager },
    { label: 'oidc', url: '/admin/settings/oidc-provider', icon: <KeyRound />, permission: AdminPermission.SiteManager },
    { label: 'security', url: '/admin/settings/security', icon: <ShieldCheck />, permission: AdminPermission.SiteManager },
    { label: 'posting', url: '/admin/settings/posting', icon: <FileText />, permission: AdminPermission.SiteManager },
    { label: 'announcement', url: '/admin/settings/announcement', icon: <Megaphone />, permission: AdminPermission.PageManager },
    { label: 'httpNotify', url: '/admin/settings/http-notify', icon: <Webhook />, permission: AdminPermission.SiteManager },
    { label: 'sensitive', url: '/admin/settings/sensitive-words', icon: <ShieldAlert />, permission: AdminPermission.SiteManager },
  ] },
]
