import type { LayoutPayload } from '@gooseforum/client'
import { hasAnyAdminPermission } from '../access'
import { adminNavGroups } from '../nav'
import type { AdminTextKey } from '../i18n'
import { shouldUseClientNavigation } from '../navigation'

export function AppSidebar({ layout, pathname, text, onNavigate, onPrefetch }: {
  layout: LayoutPayload
  pathname: string
  text(key: AdminTextKey): string
  onNavigate(path: string): void
  onPrefetch(path: string): void
}) {
  const groups = adminNavGroups.map((group) => ({
    ...group,
    items: group.items.filter((item) => hasAnyAdminPermission(layout.viewer.adminPermissions, item.permission)),
  })).filter((group) => group.items.length)

  return <nav className="flex flex-col gap-2">{groups.map(group => (
    <section key={group.label || 'main'}>
      {group.label && <h2 className="mb-1 px-2 text-xs font-medium text-muted-foreground">{text(group.label)}</h2>}
      <ul className="flex flex-col gap-0.5">{group.items.map(item => (
        <li key={item.url}><a
          href={item.url}
          aria-current={pathname === item.url ? 'page' : undefined}
          className="sidebar-nav-link flex min-h-8 items-center gap-2 rounded-md px-2 text-[13px] font-medium text-foreground/75 hover:bg-accent [&_svg]:size-4 [&_svg]:shrink-0"
          onMouseEnter={() => onPrefetch(item.url)}
          onFocus={() => onPrefetch(item.url)}
          onClick={event => {
            if (!shouldUseClientNavigation(event)) return
            event.preventDefault()
            onNavigate(item.url)
          }}
        >{item.icon}<span>{text(item.label)}</span></a></li>
      ))}</ul>
    </section>
  ))}</nav>
}
