import {
  BookOpenText,
  CircleGauge,
  FolderTree,
  HandHeart,
  LayoutDashboard,
  Settings2,
  UsersRound,
  type LucideIcon,
} from '@lucide/vue'

export interface AdminNavItem {
  label: string
  href: string
  icon: LucideIcon
  active?: boolean
}

export interface AdminNavGroup {
  label: string
  items: AdminNavItem[]
}

export const adminNavGroups: AdminNavGroup[] = [
  {
    label: 'Overview',
    items: [
      { label: '控制台', href: '/manage', icon: LayoutDashboard, active: true },
      { label: '实时概览', href: '/manage/metrics', icon: CircleGauge },
    ],
  },
  {
    label: 'Content',
    items: [
      { label: '帖子管理', href: '/manage/posts', icon: BookOpenText },
      { label: '分类管理', href: '/manage/categories', icon: FolderTree },
      { label: '用户管理', href: '/manage/users', icon: UsersRound },
    ],
  },
  {
    label: 'System',
    items: [
      { label: '赞助管理', href: '/manage/sponsors', icon: HandHeart },
      { label: '站点设置', href: '/manage/settings', icon: Settings2 },
    ],
  },
]
