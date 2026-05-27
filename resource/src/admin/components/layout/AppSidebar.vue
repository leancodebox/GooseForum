<script lang="ts" setup>
import {
  Award,
  FileText,
  ExternalLink,
  GalleryVerticalEnd,
  Heart,
  Link,
  ListChecks,
  Mail,
  Megaphone,
  Monitor,
  PanelsTopLeft,
  ShieldCheck,
  Tags,
  UserCog,
} from '@lucide/vue'
import { computed } from 'vue'
import { Avatar, AvatarFallback, AvatarImage } from '@/admin/components/ui/avatar'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/admin/components/ui/sidebar'
import { RouterLink, useRoute } from 'vue-router'
import type { LayoutPayload } from '@/types/payload'
import type { LucideIcon } from '@lucide/vue'

defineProps<{
  layout: LayoutPayload
}>()

const route = useRoute()
const currentPath = computed(() => route.path.replace(/\/+$/, '') || '/admin')

interface NavItem {
  title: string
  url: string
  icon: LucideIcon
  active?: boolean
  external?: boolean
  items?: NavItem[]
}

interface NavGroup {
  title: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    title: 'GooseForum',
    items: [
      { title: '站点统计', url: '/admin', icon: Monitor },
      { title: '用户管理', url: '/admin/users', icon: UserCog },
      { title: '角色管理', url: '/admin/roles', icon: ShieldCheck },
      { title: '分类管理', url: '/admin/categories', icon: Tags },
      { title: '帖子管理', url: '/admin/posts', icon: FileText },
      { title: '友情链接', url: '/admin/links', icon: Link },
      { title: '赞助管理', url: '/admin/sponsors', icon: Heart },
      { title: '徽章管理', url: '/admin/badges', icon: Award },
      { title: '操作记录', url: '/admin/opt-records', icon: ListChecks },
    ],
  },
  {
    title: '站点设置',
    items: [
      { title: '基础信息', url: '/admin/settings/site-info', icon: PanelsTopLeft },
      { title: '发信服务', url: '/admin/settings/mail', icon: Mail },
      { title: '安全与注册', url: '/admin/settings/security', icon: ShieldCheck },
      { title: '发布内容', url: '/admin/settings/posting', icon: FileText },
      { title: '系统公告', url: '/admin/settings/announcement', icon: Megaphone },
    ],
  },
]

function isActive(item: NavItem) {
  return currentPath.value === item.url
}
</script>

<template>
  <Sidebar collapsible="icon" class="z-50">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            as-child
            size="lg"
            class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
          >
            <a href="/" target="_blank" rel="noopener noreferrer" title="打开前台首页">
              <div class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <GalleryVerticalEnd class="size-4" />
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-semibold">{{ layout.site.name || 'GooseForum' }}</span>
                <span class="truncate text-xs">Admin</span>
              </div>
            </a>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>

    <SidebarContent class="gap-0">
      <SidebarGroup v-for="group in navGroups" :key="group.title" class="py-1">
        <SidebarGroupLabel class="h-7">{{ group.title }}</SidebarGroupLabel>
        <SidebarMenu>
          <SidebarMenuItem v-for="item in group.items" :key="item.title">
            <SidebarMenuButton as-child :is-active="isActive(item)" :tooltip="item.title">
              <a
                v-if="item.external"
                :href="item.url"
                target="_blank"
                rel="noopener noreferrer"
              >
                <component :is="item.icon" />
                <span>{{ item.title }}</span>
                <ExternalLink class="ml-auto size-4" />
              </a>
              <RouterLink v-else :to="item.url">
                <component :is="item.icon" />
                <span>{{ item.title }}</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg">
            <Avatar class="size-8 rounded-lg">
              <AvatarImage :src="layout.viewer.avatarUrl" :alt="layout.viewer.username || 'admin'" />
              <AvatarFallback class="rounded-lg">CN</AvatarFallback>
            </Avatar>
            <div class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-semibold">{{ layout.viewer.username || 'shadcn' }}</span>
              <span class="truncate text-xs">{{ layout.viewer.email || '未设置邮箱' }}</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
