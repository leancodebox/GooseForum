import type { AuthLocale } from '@gooseforum/runtime/i18n/auth'
import { Button } from '@gooseforum/ui/components/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuTrigger } from '@gooseforum/ui/components/dropdown-menu'
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
  zh: { language: '切换语言', light: '切换到浅色主题', dark: '切换到深色主题', site: '返回站点' },
  en: { language: 'Switch language', light: 'Switch to light theme', dark: 'Switch to dark theme', site: 'Back to site' },
  ja: { language: '言語を切り替え', light: 'ライトテーマに切り替え', dark: 'ダークテーマに切り替え', site: 'サイトに戻る' },
  it: { language: 'Cambia lingua', light: 'Passa al tema chiaro', dark: 'Passa al tema scuro', site: 'Torna al sito' },
} as const

export function SiteHeader({ title, locale, theme, onLocaleChange, onThemeToggle }: {
  title: string
  locale: AuthLocale
  theme: 'gf-light' | 'gf-dark'
  onLocaleChange(locale: AuthLocale): void
  onThemeToggle(): void
}) {
  const labels = headerLabels[locale]
  const themeLabel = theme === 'gf-dark' ? labels.light : labels.dark

  return <header className="sticky top-0 z-40 flex h-(--header-height) shrink-0 items-center gap-1 border-b bg-background/92 px-3 text-foreground backdrop-blur transition-[width,height] ease-linear supports-[backdrop-filter]:bg-background/80 lg:gap-2 lg:px-4">
    <SidebarTrigger className="-ml-1 size-8 shrink-0" />
    <Separator orientation="vertical" className="mx-2 h-4! self-center!" />
    <h1 className="min-w-0 flex-1 truncate text-base font-medium">{title}</h1>
    <div className="ml-auto flex shrink-0 items-center gap-1">
      <Button variant="ghost" size="icon-sm" title={themeLabel} aria-label={themeLabel} onClick={onThemeToggle}>
        {theme === 'gf-dark' ? <Sun /> : <Moon />}
      </Button>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild><Button variant="ghost" size="icon-sm" title={labels.language} aria-label={labels.language}><Languages /></Button></DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-36"><DropdownMenuGroup>{(Object.keys(localeLabels) as AuthLocale[]).map((item) => <DropdownMenuItem key={item} data-current={locale === item} className="data-[current=true]:font-semibold data-[current=true]:text-primary" onSelect={() => onLocaleChange(item)}>{localeLabels[item]}</DropdownMenuItem>)}</DropdownMenuGroup></DropdownMenuContent>
      </DropdownMenu>
      <Button asChild variant="ghost" size="sm" className="hidden xl:inline-flex"><a href="/"><ArrowLeft data-icon="inline-start" />{labels.site}</a></Button>
    </div>
  </header>
}
