import { Header } from './header'
import { Main } from './main'
import { TopNav } from './top-nav'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ConfigDrawer } from '@/components/config-drawer'
import { ThemeSwitch } from '@/components/theme-switch'
import { Separator } from '@/components/ui/separator'

interface ContentLayoutProps {
  title: string
  description?: string
  children: React.ReactNode
  headerActions?: React.ReactNode
  showSeparator?: boolean
  topNav?: {
    title: string
    href: string
    isActive: boolean
    disabled?: boolean
  }[]
}

export function ContentLayout({
  title,
  description,
  children,
  headerActions,
  showSeparator = false,
  topNav,
}: ContentLayoutProps) {
  return (
    <>
      <Header fixed>
        {topNav ? <TopNav links={topNav} /> : <Search />}
        <div className='ms-auto flex items-center space-x-4'>
          {topNav && <Search />}
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-col gap-4'>
        <div className='flex items-center justify-between'>
          <div className='space-y-0.5'>
            <h2 className='text-2xl font-bold tracking-tight'>{title}</h2>
            {description && (
              <p className='text-muted-foreground'>{description}</p>
            )}
          </div>
          {headerActions && <div>{headerActions}</div>}
        </div>
        {showSeparator && <Separator className='my-2' />}
        {children}
      </Main>
    </>
  )
}
