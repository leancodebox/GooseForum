import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import type { GooseAdminApi } from '@gooseforum/client'
import { createAccessGroupText } from '../access-groups-i18n'
import { prepareAdminTranslations } from '../translation-loader'
import { AccessGroupsManagementPage } from './access-groups-management-page'

afterEach(cleanup)

it('uses translated system names in the list, details and search, and preserves custom names', async () => {
  const groups = [
    ['trust:100', '成员及以上'], ['trust:200', '稳定成员及以上'],
    ['trust:1000', '社区骨干及以上'], ['', '社区骨干及以上'],
  ].map(([systemKey, name], id) => ({ id: id + 1, systemKey, name, members: [], grants: [], status: 1 }))
  const overview = vi.fn().mockResolvedValue({ groups, categories: [] })
  await prepareAdminTranslations('en', ["admin.access-groups"])
  const api = { accessGroups: { overview } } as unknown as GooseAdminApi
  const onNavigate = vi.fn()
  const { rerender } = render(<AccessGroupsManagementPage api={api} text={createAccessGroupText('en')} onNavigate={onNavigate} />)
  const list = within(await screen.findByRole('navigation', { name: 'Groups' }))
  expect(list.getByText('Member and above')).toBeTruthy()
  expect(list.getByText('Community leader and above')).toBeTruthy()
  expect(list.getByText('社区骨干及以上')).toBeTruthy()
  await userEvent.type(screen.getByRole('textbox', { name: 'Groups' }), 'regular')
  expect(list.getAllByRole('button')).toHaveLength(1)
  await userEvent.click(list.getByText('Regular and above'))
  expect(screen.getByRole('heading', { name: 'Regular and above' })).toBeTruthy()
  await prepareAdminTranslations('zh', ["admin.access-groups"])
  rerender(<AccessGroupsManagementPage api={api} text={createAccessGroupText('zh')} onNavigate={onNavigate} />)
  expect(screen.getByRole('heading', { name: '稳定成员及以上' })).toBeTruthy()
  expect(overview).toHaveBeenCalledTimes(1)
})
