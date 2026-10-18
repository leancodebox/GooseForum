import type { ReactNode } from 'react'
import { ShieldCheck, FileText, BadgeCheck, KeyRound, Fingerprint } from 'lucide-react'
import { canVisitAdminPath } from '../access'
import type { AdminTextKey } from '../i18n'
import { shouldUseClientNavigation } from '../navigation'

import { communityRulePaths, identitySettingPaths } from '../navigation'

export function SettingsSectionLayout({ pathname, permissions, text, navigate, children, section }: {
  section: 'community' | 'identity'
  pathname: string
  permissions: number[]
  text(key: AdminTextKey): string
  navigate(path: string): void
  children: ReactNode
}) {
  const items = (section === 'identity' ? [
    { path: identitySettingPaths[0], label: 'oauthLogin' as const, icon: KeyRound },
    { path: identitySettingPaths[1], label: 'oidcProvider' as const, icon: Fingerprint },
  ] : [
    { path: communityRulePaths[0], label: 'security' as const, icon: ShieldCheck },
    { path: communityRulePaths[1], label: 'posting' as const, icon: FileText },
    { path: communityRulePaths[2], label: 'trust' as const, icon: BadgeCheck },
  ]).filter(item => canVisitAdminPath(permissions, item.path))
  return <div className={`grid min-w-0 gap-5 px-3 py-4 lg:gap-7 lg:px-4 ${section === 'identity' ? 'lg:grid-cols-[12rem_minmax(0,1fr)]' : 'lg:grid-cols-[10rem_minmax(0,1fr)]'}`}>
    <nav aria-label={text(section === 'identity' ? 'oauthOidc' : 'communityRules')} className="flex flex-wrap content-start gap-0.5 border-b pb-3 lg:flex-col lg:flex-nowrap lg:items-stretch lg:content-normal lg:border-b-0 lg:border-r lg:pr-3 lg:pb-0">
      {items.map(({ path, label, icon: Icon }) => <a key={path} href={path} title={text(label)} aria-current={pathname === path ? 'page' : undefined}
        className="flex h-8 min-w-0 shrink-0 items-center gap-2 rounded-md px-2 py-0 text-[13px] leading-[18.5714px] text-foreground/75 hover:bg-accent hover:text-foreground lg:w-full pointer-coarse:h-11 aria-[current=page]:bg-primary/10 aria-[current=page]:font-medium aria-[current=page]:text-primary aria-[current=page]:hover:bg-primary/15 aria-[current=page]:hover:text-primary aria-[current=page]:[&>svg]:text-primary"
        onClick={event => { if (shouldUseClientNavigation(event)) { event.preventDefault(); navigate(path) } }}><Icon className="size-4 shrink-0" /><span className="min-w-0 truncate">{text(label)}</span></a>)}
    </nav>
    <div className="min-w-0 [&>main]:p-0">{children}</div>
  </div>
}
