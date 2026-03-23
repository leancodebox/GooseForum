import { useLayout } from '@/context/layout-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { sidebarData } from './data/sidebar-data'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { TeamSwitcher } from './team-switcher'
import { useAuthStore } from '@/stores/auth-store'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const { auth } = useAuthStore()

  const isAdmin = auth.user?.role.includes('admin')
  const userEmail = auth.user?.email || ''

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <TeamSwitcher
          teams={[
            {
              name: 'GooseForum Admin',
              plan: isAdmin ? 'Administrator' : 'User',
            },
          ]}
        />
      </SidebarHeader>
      <SidebarContent>
        {sidebarData.navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser
          user={{
            name: auth.user?.accountNo || 'Guest',
            email: userEmail,
            avatar: '/static/pic/2.webp',
          }}
        />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
