import { ShieldCheck } from 'lucide-react'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

type TeamSwitcherProps = {
  teams: {
    name: string
    plan: string
  }[]
}

export function TeamSwitcher({ teams }: TeamSwitcherProps) {
  const activeTeam = teams[0]

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          size='lg'
          asChild
          className='hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
        >
          <a href='/'>
            <div className='flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground'>
              <ShieldCheck className='size-5' />
            </div>
            <div className='grid flex-1 text-start text-sm leading-tight'>
              <span className='truncate font-bold tracking-tight text-base'>
                {activeTeam?.name || 'GooseForum'}
              </span>
              <span className='truncate text-[10px] uppercase tracking-widest text-muted-foreground font-semibold'>
                {activeTeam?.plan || 'Admin'}
              </span>
            </div>
          </a>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
