<script lang="ts" setup>
import { adminText } from '@/admin/runtime/i18n-text'
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
  Webhook,
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
import { useI18n } from 'vue-i18n'
import { AdminPermission, hasAdminPermission } from '@/admin/runtime/access'

defineProps<{
  layout: LayoutPayload
}>()

const route = useRoute()
const { locale } = useI18n()
const currentPath = computed(() => route.path.replace(/\/+$/, '') || '/admin')

interface NavItem {
  title: string
  url: string
  icon: LucideIcon
  permission?: AdminPermission
  active?: boolean
  external?: boolean
  items?: NavItem[]
}

interface NavGroup {
  title: string
  items: NavItem[]
}

const navGroups = computed<NavGroup[]>(() => {
  locale.value
  return [
  {
    title: 'GooseForum',
    items: [
      { title: adminText('k004c'), url: '/admin', icon: Monitor, permission: AdminPermission.Admin },
      { title: adminText('k006i'), url: '/admin/users', icon: UserCog, permission: AdminPermission.UserManager },
      { title: adminText('k007f'), url: '/admin/roles', icon: ShieldCheck, permission: AdminPermission.RoleManager },
      { title: adminText('k005l'), url: '/admin/categories', icon: Tags, permission: AdminPermission.ArticlesManager },
      { title: adminText('k005u'), url: '/admin/posts', icon: FileText, permission: AdminPermission.ArticlesManager },
      { title: adminText('k002j'), url: '/admin/links', icon: Link, permission: AdminPermission.PageManager },
      { title: adminText('k004o'), url: '/admin/sponsors', icon: Heart, permission: AdminPermission.PageManager },
      { title: adminText('k0058'), url: '/admin/badges', icon: Award, permission: AdminPermission.SiteManager },
      { title: adminText('k007c'), url: '/admin/opt-records', icon: ListChecks, permission: AdminPermission.Admin },
    ],
  },
  {
    title: adminText('k007t'),
    items: [
      { title: adminText('k007u'), url: '/admin/settings/site-info', icon: PanelsTopLeft, permission: AdminPermission.SiteManager },
      { title: adminText('k007v'), url: '/admin/settings/mail', icon: Mail, permission: AdminPermission.SiteManager },
      { title: adminText('k0005'), url: '/admin/settings/security', icon: ShieldCheck, permission: AdminPermission.SiteManager },
      { title: adminText('k007w'), url: '/admin/settings/posting', icon: FileText, permission: AdminPermission.SiteManager },
      { title: adminText('k0009'), url: '/admin/settings/announcement', icon: Megaphone, permission: AdminPermission.PageManager },
      { title: adminText('k00cj'), url: '/admin/settings/http-notify', icon: Webhook, permission: AdminPermission.SiteManager },
    ],
  },
  ].map(group => ({
    ...group,
    items: group.items.filter(item => item.permission === undefined || hasAdminPermission(item.permission)),
  })).filter(group => group.items.length > 0)
})

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
            <a href="/" target="_blank" rel="noopener noreferrer" :title="adminText('k007s')">
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
              <span class="truncate text-xs">{{ layout.viewer.email || adminText('k007x') }}</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
