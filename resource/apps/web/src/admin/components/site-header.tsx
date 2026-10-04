import type { AuthLocale } from '@gooseforum/runtime/i18n/auth'
import type { BrowserThemePreference } from '@gooseforum/runtime/browser-host'
import { Button } from '@gooseforum/ui/components/button'
import { RoutineIcon } from '@gooseforum/ui/components/icons/material-symbols/routine-icon'
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuTrigger } from '@gooseforum/ui/components/dropdown-menu'
import { Separator } from '@gooseforum/ui/components/separator'
import { SidebarTrigger } from '@gooseforum/ui/components/sidebar'
import { ArrowLeft, Languages, Moon, Sun } from 'lucide-react'

const localeLabels: Record<AuthLocale, string> = {
  zh: '简体中文',
  en: 'English',
  ja: '日本語',
  it: 'Italiano',
}

const headerLabels = {
  zh: { language: '切换语言', theme: '选择主题', light: '浅色', dark: '深色', system: '跟随系统', site: '返回站点' },
  en: { language: 'Switch language', theme: 'Choose theme', light: 'Light', dark: 'Dark', system: 'System', site: 'Back to site' },
  ja: { language: '言語を切り替え', theme: 'テーマを選択', light: 'ライト', dark: 'ダーク', system: 'システム', site: 'サイトに戻る' },
  it: { language: 'Cambia lingua', theme: 'Scegli tema', light: 'Chiaro', dark: 'Scuro', system: 'Sistema', site: 'Torna al sito' },
} as const

export function SiteHeader({ title, locale, themePreference, onLocaleChange, onThemeChange }: {
  title: string
  locale: AuthLocale
  themePreference: BrowserThemePreference
  onLocaleChange(locale: AuthLocale): void
  onThemeChange(preference: BrowserThemePreference): void
}) {
  const labels = headerLabels[locale]

  return <header className="sticky top-0 z-40 flex h-(--header-height) shrink-0 items-center gap-1 border-b bg-background/92 px-3 text-foreground backdrop-blur transition-[width,height] ease-linear supports-[backdrop-filter]:bg-background/80 lg:gap-2 lg:px-4">
    <SidebarTrigger className="-ml-1 size-8 shrink-0" />
    <Separator orientation="vertical" className="mx-2 h-4! self-center!" />
    <h1 className="min-w-0 flex-1 truncate text-base font-medium">{title}</h1>
    <div className="ml-auto flex shrink-0 items-center gap-1">
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild><Button variant="ghost" size="icon-sm" className="[&_svg]:size-5" title={labels.theme} aria-label={labels.theme}>
          {themePreference === 'system' ? <RoutineIcon /> : themePreference === 'gf-light' ? <Sun /> : <Moon />}
        </Button></DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-36">
          <DropdownMenuRadioGroup value={themePreference} onValueChange={(value) => {
            if (value === 'system' || value === 'gf-light' || value === 'gf-dark') onThemeChange(value)
          }}>
            <DropdownMenuRadioItem value="gf-light">{labels.light}</DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="gf-dark">{labels.dark}</DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="system">{labels.system}</DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild><Button variant="ghost" size="icon-sm" title={labels.language} aria-label={labels.language}><Languages /></Button></DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-36"><DropdownMenuGroup>{(Object.keys(localeLabels) as AuthLocale[]).map((item) => <DropdownMenuItem key={item} data-current={locale === item} className="data-[current=true]:font-semibold data-[current=true]:text-primary" onSelect={() => onLocaleChange(item)}>{localeLabels[item]}</DropdownMenuItem>)}</DropdownMenuGroup></DropdownMenuContent>
      </DropdownMenu>
      <Button asChild variant="ghost" size="sm" className="hidden xl:inline-flex"><a href="/"><ArrowLeft data-icon="inline-start" />{labels.site}</a></Button>
    </div>
  </header>
}
