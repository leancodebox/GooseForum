import { Search } from '@/components/search'
import { ConfigDrawer } from '@/components/config-drawer'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Skeleton } from '@/components/ui/skeleton'

export function RoutePending() {
  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-col gap-4'>
        <div className='flex items-center justify-between'>
          <div className='space-y-2'>
            <Skeleton className='h-7 w-40' />
            <Skeleton className='h-4 w-72 max-w-[70vw]' />
          </div>
          <Skeleton className='hidden h-9 w-24 sm:block' />
        </div>

        <div className='rounded-lg border bg-card'>
          <div className='space-y-3 p-4'>
            <Skeleton className='h-9 w-full max-w-sm' />
            <div className='space-y-2'>
              {Array.from({ length: 6 }).map((_, index) => (
                <div
                  key={index}
                  className='grid grid-cols-[1fr_auto] items-center gap-4 rounded-md border px-4 py-3'
                >
                  <div className='space-y-2'>
                    <Skeleton className='h-4 w-full max-w-xl' />
                    <Skeleton className='h-3 w-2/3 max-w-md' />
                  </div>
                  <Skeleton className='h-8 w-8 rounded-full' />
                </div>
              ))}
            </div>
          </div>
        </div>
      </Main>
    </>
  )
}
