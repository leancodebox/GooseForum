import {
  Link,
  Construction,
  LayoutDashboard,
  Monitor,
  Bug,
  ListTodo,
  FileX,
  HelpCircle,
  Lock,
  Bell,
  Package,
  Palette,
  ServerOff,
  Settings,
  PanelsTopLeft,
  Wrench,
  UserCog,
  UserX,
  Users,
  MessagesSquare,
  ShieldCheck,
  FileText,
  Tags,
  Mail,
  PanelBottom,
  Heart,
  LifeBuoy,
  AudioWaveform,
  Command,
  GalleryVerticalEnd,
} from 'lucide-react'
import { type SidebarData } from '../types'

const allNavGroups = [
  {
    title: 'GooseForum',
    items: [
      {
        title: '站点统计',
        url: '/',
        icon: Monitor,
      },
      {
        title: '用户管理',
        url: '/users-management',
        icon: UserCog,
      },
      {
        title: '角色管理',
        url: '/roles-management',
        icon: ShieldCheck,
      },
      {
        title: '分类管理',
        url: '/categories-management',
        icon: Tags,
      },
      {
        title: '帖子管理',
        url: '/posts-management',
        icon: FileText,
      },
      {
        title: '友情链接',
        url: '/links-management',
        icon: Link,
      },
      {
        title: '赞助管理',
        url: '/sponsorship-management',
        icon: Heart,
      },
      {
        title: '站点设置',
        icon: Settings,
        items: [
          {
            title: '站点信息',
            url: '/site-settings/site-info',
            icon: PanelsTopLeft,
          },
          {
            title: '页脚管理',
            url: '/site-settings/footer-management',
            icon: PanelBottom,
          },
          {
            title: '邮箱设置',
            url: '/site-settings/mail-settings',
            icon: Mail,
          },
          {
            title: '安全与注册',
            url: '/site-settings/security-settings',
            icon: ShieldCheck,
          },
          {
            title: '发布内容',
            url: '/site-settings/posting-settings',
            icon: FileText,
          },
        ],
      },
    ],
  },
  {
    title: 'Shadcn',
    items: [
      {
        title: 'Dashboard',
        url: '/dashboard-stats',
        icon: LayoutDashboard,
      },
      {
        title: 'Tasks',
        url: '/tasks',
        icon: ListTodo,
      },
      {
        title: 'Apps',
        url: '/apps',
        icon: Package,
      },
      {
        title: 'Chats',
        url: '/chats',
        badge: '3',
        icon: MessagesSquare,
      },
      {
        title: 'Users',
        url: '/users',
        icon: Users,
      },
    ],
  },
  {
    title: 'Pages',
    items: [
      {
        title: 'Auth',
        icon: ShieldCheck,
        items: [
          {
            title: 'Sign In',
            url: '/sign-in',
          },
          {
            title: 'Sign In (2 Col)',
            url: '/sign-in-2',
          },
          {
            title: 'Sign Up',
            url: '/sign-up',
          },
          {
            title: 'Forgot Password',
            url: '/forgot-password',
          },
          {
            title: 'OTP',
            url: '/otp',
          },
        ],
      },
      {
        title: 'Errors',
        icon: Bug,
        items: [
          {
            title: 'Unauthorized',
            url: '/errors/unauthorized',
            icon: Lock,
          },
          {
            title: 'Forbidden',
            url: '/errors/forbidden',
            icon: UserX,
          },
          {
            title: 'Not Found',
            url: '/errors/not-found',
            icon: FileX,
          },
          {
            title: 'Internal Server Error',
            url: '/errors/internal-server-error',
            icon: ServerOff,
          },
          {
            title: 'Maintenance Error',
            url: '/errors/maintenance-error',
            icon: Construction,
          },
        ],
      },
    ],
  },
  {
    title: 'Other',
    items: [
      {
        title: 'Settings',
        icon: Settings,
        items: [
          {
            title: 'Profile',
            url: '/settings',
            icon: UserCog,
            },
          {
            title: 'Account',
            url: '/settings/account',
            icon: Wrench,
          },
          {
            title: 'Appearance',
            url: '/settings/appearance',
            icon: Palette,
          },
          {
            title: 'Notifications',
            url: '/settings/notifications',
            icon: Bell,
          },
          {
            title: 'Display',
            url: '/settings/display',
            icon: Monitor,
          },
        ],
      },
      {
        title: 'Help Center',
        url: '/help-center',
        icon: HelpCircle,
      },
    ],
  },
]

export const sidebarData: SidebarData = {
  user: {
    name: 'satnaing',
    email: 'satnaingdev@gmail.com',
    avatar: '/static/pic/2.webp',
  },
  teams: [
    {
      name: 'GooseForum Admin',
      logo: Command,
      plan: 'Vite + ShadcnUI',
    },
    {
      name: 'Acme Inc',
      logo: GalleryVerticalEnd,
      plan: 'Enterprise',
    },
    {
      name: 'Acme Corp.',
      logo: AudioWaveform,
       plan: 'Startup',
    },
  ],
  navGroups: import.meta.env.DEV
    ? allNavGroups
    : allNavGroups.filter((group) => group.title === 'GooseForum'),
}
