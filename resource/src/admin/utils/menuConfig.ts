import type { Component } from 'vue'
import {
  BookOpenIcon,
  DocumentDuplicateIcon,
  DocumentTextIcon,
  FolderIcon,
  GiftIcon,
  GlobeAltIcon,
  HomeIcon,
  LinkIcon,
  ShieldCheckIcon,
  TagIcon,
  TicketIcon,
  UsersIcon,
  DocumentMinusIcon,
  Cog6ToothIcon
} from '@heroicons/vue/24/outline'

export interface MenuItem {
  key: string
  label: string
  path?: string
  icon: Component
  children?: MenuItem[]
}

export interface MenuItemResult {
  item: MenuItem
  parent: MenuItem | null
}

export const menuItems: MenuItem[] = [
  {
    key: 'dashboard',
    label: '仪表盘',
    path: '/admin',
    icon: HomeIcon
  },
  {
    key: 'users',
    label: '用户管理',
    path: '/admin/users',
    icon: UsersIcon
  },
  {
    key: 'roles',
    label: '角色管理',
    path: '/admin/roles',
    icon: ShieldCheckIcon
  },
  {
    key: 'posts',
    label: '帖子管理',
    path: '/admin/posts',
    icon: DocumentTextIcon
  },
  {
    key: 'categories',
    label: '分类管理',
    path: '/admin/categories',
    icon: TagIcon
  },
  {
    key: 'friend-links',
    label: '友情链接',
    path: '/admin/friend-links',
    icon: LinkIcon
  },
  {
    key: 'footer-links',
    label: '页脚管理',
    path: '/admin/footer-links',
    icon: DocumentMinusIcon
  },
  {
    key: 'sponsors',
    label: '赞助管理',
    path: '/admin/sponsors',
    icon: GiftIcon
  },
  {
    key: 'tickets',
    label: '工单管理',
    icon: TicketIcon,
    path: '/admin/tickets/view',
  },
  {
    key: 'docs',
    label: '文档管理',
    icon: FolderIcon,
    children: [
      {
        key: 'docs-projects',
        label: '项目管理',
        path: '/admin/docs/projects',
        icon: BookOpenIcon
      },
      {
        key: 'docs-versions',
        label: '版本管理',
        path: '/admin/docs/versions',
        icon: DocumentDuplicateIcon
      },
      {
        key: 'docs-contents',
        label: '内容管理',
        path: '/admin/docs/contents',
        icon: DocumentTextIcon
      }
    ]
  },
  {
    key: 'site-settings',
    label: '站点设置',
    path: '/admin/site-settings',
    icon: Cog6ToothIcon
  },
]
