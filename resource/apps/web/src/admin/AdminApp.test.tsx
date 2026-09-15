import type { ReactNode } from 'react'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { AdminApp } from './AdminApp'
import { prepareAdminPage } from './page-registry'

vi.mock('./page-registry', () => ({
  prepareAdminPage: vi.fn(),
    DashboardPage: () => <p>DashboardPage</p>,
    CategoriesManagementPage: () => <p>CategoriesManagementPage</p>,
    AccessGroupsManagementPage: () => <p>AccessGroupsManagementPage</p>,
    RolesManagementPage: () => <p>RolesManagementPage</p>,
    UsersManagementPage: () => <p>UsersManagementPage</p>,
    PostsManagementPage: () => <p>PostsManagementPage</p>,
    LinksManagementPage: () => <p>LinksManagementPage</p>,
    SponsorsManagementPage: () => <p>SponsorsManagementPage</p>,
    BadgesManagementPage: () => <p>BadgesManagementPage</p>,
    FileResourcesManagementPage: () => <p>FileResourcesManagementPage</p>,
    OptRecordsManagementPage: () => <p>OptRecordsManagementPage</p>,
    SiteInfoManagementPage: () => <p>SiteInfoManagementPage</p>,
    SiteChromeManagementPage: () => <p>SiteChromeManagementPage</p>,
    MailSettingsPage: () => <p>MailSettingsPage</p>,
    SecuritySettingsPage: () => <p>SecuritySettingsPage</p>,
    PostingSettingsPage: () => <p>PostingSettingsPage</p>,
    AnnouncementSettingsPage: () => <p>AnnouncementSettingsPage</p>,
    HttpNotifySettingsPage: () => <p>HttpNotifySettingsPage</p>,
    SensitiveWordsSettingsPage: () => <p>SensitiveWordsSettingsPage</p>,
    OAuthSettingsPage: () => <p>OAuthSettingsPage</p>,
    OIDCProviderSettingsPage: () => <p>OIDCProviderSettingsPage</p>,
}))
vi.mock('./components/app-sidebar', () => ({ AppSidebar: ({ onNavigate }: { onNavigate(path: string): void }) => <nav><button onClick={() => onNavigate('/admin/users')}>Users</button><button onClick={() => onNavigate('/admin/sponsors')}>Sponsors</button></nav> }))
vi.mock('./components/site-header', () => ({ SiteHeader: () => null }))
vi.mock('@gooseforum/ui/components/sidebar', () => ({ SidebarProvider: ({ children }: { children: ReactNode }) => <div>{children}</div>, SidebarInset: ({ children }: { children: ReactNode }) => <main>{children}</main> }))
vi.mock('@gooseforum/ui/components/tooltip', () => ({ TooltipProvider: ({ children }: { children: ReactNode }) => <>{children}</> }))
vi.mock('@gooseforum/ui/components/sonner', () => ({ Toaster: () => null }))
vi.mock('../host/browser-runtime', () => ({ detectBrowserLocale: () => 'zh', detectBrowserTheme: () => 'gf-light', applyBrowserLocale: vi.fn(), applyBrowserTheme: vi.fn() }))

const prepare = vi.mocked(prepareAdminPage)
const page = { layout: { site: { name: 'GooseForum' }, theme: { colors: {} } } } as PagePayload
function mount() { render(<AdminApp page={page} api={{} as GooseAdminApi} />) }
function deferred() { let resolve!: () => void; const promise = new Promise<void>(done => { resolve = done }); return { promise, resolve } }
beforeEach(() => { window.history.replaceState(null, '', '/admin'); prepare.mockReset() })
afterEach(cleanup)

describe('prepared admin navigation', () => {
  it('retains the current page and URL until the target module is ready', async () => {
    const load = deferred(); prepare.mockReturnValue(load.promise); mount()
    fireEvent.click(screen.getByText('Users'))
    expect(screen.getByText('DashboardPage')).toBeTruthy()
    expect(screen.queryByText('UsersManagementPage')).toBeNull()
    expect(window.location.pathname).toBe('/admin')
    await act(async () => { load.resolve(); await load.promise })
    expect(screen.getByText('UsersManagementPage')).toBeTruthy()
    expect(window.location.pathname).toBe('/admin/users')
  })
  it('does not let an older module load override the latest menu selection', async () => {
    const users = deferred(); const sponsors = deferred(); prepare.mockImplementation(path => path === '/admin/users' ? users.promise : sponsors.promise); mount()
    fireEvent.click(screen.getByText('Users')); fireEvent.click(screen.getByText('Sponsors'))
    await act(async () => { sponsors.resolve(); await sponsors.promise })
    await act(async () => { users.resolve(); await users.promise })
    expect(screen.getByText('SponsorsManagementPage')).toBeTruthy()
    expect(screen.queryByText('UsersManagementPage')).toBeNull()
    expect(window.location.pathname).toBe('/admin/sponsors')
  })
  it('preserves the current page on failure and supports retrying the menu', async () => {
    prepare.mockRejectedValueOnce(new Error('Module unavailable')).mockResolvedValue(undefined); mount()
    fireEvent.click(screen.getByText('Users'))
    expect(await screen.findByRole('alert')).toHaveProperty('textContent', 'Module unavailable')
    expect(screen.getByText('DashboardPage')).toBeTruthy()
    fireEvent.click(screen.getByText('Users'))
    await waitFor(() => expect(screen.getByText('UsersManagementPage')).toBeTruthy())
    expect(screen.queryByRole('alert')).toBeNull()
  })
  it('also prepares browser history destinations before replacing the rendered page', async () => {
    const load = deferred(); prepare.mockReturnValue(load.promise); mount()
    window.history.replaceState(null, '', '/admin/users')
    fireEvent(window, new PopStateEvent('popstate'))
    expect(screen.getByText('DashboardPage')).toBeTruthy()
    await act(async () => { load.resolve(); await load.promise })
    expect(screen.getByText('UsersManagementPage')).toBeTruthy()
    expect(prepare).toHaveBeenCalledWith('/admin/users')
  })
})
