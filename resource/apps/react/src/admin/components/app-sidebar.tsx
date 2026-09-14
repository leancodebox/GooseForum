import * as React from 'react'
import type { LayoutPayload } from '@gooseforum/client'
import { Avatar, AvatarFallback, AvatarImage } from '@gooseforum/react/components/ui/avatar'
import {
  Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupLabel, SidebarHeader,
  SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarRail, useSidebar,
} from '@gooseforum/react/components/ui/sidebar'
import { ExternalLink, GalleryVerticalEnd } from 'lucide-react'
import { hasAnyAdminPermission } from '../access'
import { adminNavGroups } from '../nav'
import type { AdminTextKey } from '../i18n'
import { shouldUseClientNavigation } from '../navigation'

export function AppSidebar({ layout, pathname, text, onNavigate, onPrefetch, ...props }: React.ComponentProps<typeof Sidebar> & {
  layout: LayoutPayload
  pathname: string
  text(key: AdminTextKey): string
  onNavigate(path: string): void
  onPrefetch(path: string): void
}) {
  const { setOpenMobile } = useSidebar()
  const logo = layout.site.logo || layout.site.brandImage || layout.site.favicon
  const initials = (layout.viewer.username || 'A').slice(0, 2).toUpperCase()
  const groups = adminNavGroups.map((group) => ({
    ...group,
    items: group.items.filter((item) => hasAnyAdminPermission(layout.viewer.adminPermissions, item.permission)),
  })).filter((group) => group.items.length)

  return <Sidebar collapsible="offcanvas" {...props}>
    <SidebarHeader className="px-2 pt-2"><SidebarMenu><SidebarMenuItem><SidebarMenuButton asChild size="lg" className="h-10 px-2"><a href="/" target="_blank" rel="noreferrer" title={text('openSite')}>
      <div className="flex size-8 items-center justify-center overflow-hidden rounded-full border bg-background shadow-xs">{logo ? <img src={logo} alt={layout.site.name || 'GooseForum'} className="size-full object-cover" /> : <GalleryVerticalEnd className="size-4.5" />}</div>
      <div className="grid flex-1 text-left leading-tight"><span className="truncate text-base font-semibold tracking-tight">{layout.site.name || 'GooseForum'}</span><span className="truncate text-[11px] text-muted-foreground">{text('console')}</span></div><ExternalLink className="size-3.5 text-muted-foreground" />
    </a></SidebarMenuButton></SidebarMenuItem></SidebarMenu></SidebarHeader>
    <SidebarContent className="gap-1 px-1 py-1.5">{groups.map((group, index) => <SidebarGroup key={group.label || 'main'} className="px-1.5 py-0.5">
      {index > 0 && group.label && <SidebarGroupLabel className="h-7 px-2 text-xs font-medium">{text(group.label)}</SidebarGroupLabel>}
      <SidebarMenu className="gap-0.5">{group.items.map((item) => <SidebarMenuItem key={item.url}><SidebarMenuButton asChild isActive={pathname === item.url} size="sm" tooltip={text(item.label)} className="h-7.5 rounded-md px-2 text-[13px] font-medium text-sidebar-foreground/75 data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-primary"><a
        href={item.url}
        onMouseEnter={() => onPrefetch(item.url)}
        onFocus={() => onPrefetch(item.url)}
        onClick={(event) => {
          if (!shouldUseClientNavigation(event)) return
          event.preventDefault()
          onNavigate(item.url)
          setOpenMobile(false)
        }}
      >{item.icon}<span>{text(item.label)}</span></a></SidebarMenuButton></SidebarMenuItem>)}</SidebarMenu>
    </SidebarGroup>)}</SidebarContent>
    <SidebarFooter className="px-2 pb-2"><SidebarMenu><SidebarMenuItem><SidebarMenuButton size="lg" className="h-11 rounded-xl px-2"><Avatar className="size-8"><AvatarImage className="object-cover" src={layout.viewer.avatarUrl} alt={layout.viewer.username || 'Admin'} /><AvatarFallback>{initials}</AvatarFallback></Avatar><div className="grid flex-1 text-left text-sm leading-tight"><span className="truncate font-semibold">{layout.viewer.username || 'Admin'}</span><span className="truncate text-xs text-muted-foreground">{layout.viewer.email}</span></div></SidebarMenuButton></SidebarMenuItem></SidebarMenu></SidebarFooter>
    <SidebarRail />
  </Sidebar>
}
