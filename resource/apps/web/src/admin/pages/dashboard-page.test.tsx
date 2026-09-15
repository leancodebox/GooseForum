import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import type { GooseAdminApi } from '@gooseforum/client'
import DashboardPage from './dashboard-page'

vi.mock('../components/traffic-overview', () => ({ TrafficOverview: () => <div>Traffic</div> }))
afterEach(cleanup)

it('keeps releases bounded and scrollable without stretching the card to the chart height', async () => {
  const releases = Array.from({ length: 12 }, (_, id) => ({ id, tag_name: `v1.${id}`, html_url: `https://example.com/releases/${id}`, body: 'Release notes', published_at: '2026-09-15T00:00:00Z' }))
  const api = { dashboard: { statistics: vi.fn().mockResolvedValue({}), version: vi.fn().mockResolvedValue({ version: 'dev', mode: 'development' }), releases: vi.fn().mockResolvedValue(releases), traffic: vi.fn().mockResolvedValue([]) } } as unknown as GooseAdminApi
  render(<DashboardPage api={api} text={key => key} locale="zh" />)
  const last = await screen.findByText('v1.11')
  const list = last.closest('a')!.parentElement!
  expect(list.children).toHaveLength(12)
  expect(list.classList.contains('max-h-96')).toBe(true)
  expect(list.classList.contains('overflow-y-auto')).toBe(true)
  expect(list.tabIndex).toBe(0)
  expect(list.closest('section')?.classList.contains('self-start')).toBe(true)
})
