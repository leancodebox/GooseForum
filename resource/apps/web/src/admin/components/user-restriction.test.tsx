import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { AdminUser, GooseAdminApi } from '@gooseforum/client'
import messages from '../messages/en-users'
import { UsersManagementPage } from '../pages/users-management-page'

afterEach(() => { cleanup(); vi.useRealTimers() })
vi.stubGlobal('ResizeObserver', class { observe() {} unobserve() {} disconnect() {} })

it('edits an expiring ban and displays its administrator history', async () => {
  vi.useFakeTimers({toFake:['Date']})
  vi.setSystemTime(new Date(2026, 9, 4))
  const user: AdminUser = { userId:7,username:'alice',email:'alice@example.com',restrictionStatus:'banned',restrictionReason:'Spam',validate:1,prestige:0,roleId:0,createTime:'2026-10-03' }
  const edit = vi.fn().mockResolvedValue(undefined)
  const history = vi.fn().mockResolvedValue({list:[{id:1,userId:7,actorId:1,status:'banned',reason:'Spam',createdAt:'2026-10-03T00:00:00Z'}],page:1,pageSize:10,total:1})
  const api = {users:{list:vi.fn().mockResolvedValue({list:[user],total:1}),edit,roles:vi.fn().mockResolvedValue([]),badgeOptions:vi.fn().mockResolvedValue({options:[],active:[]}),saveBadges:vi.fn().mockResolvedValue(undefined),restrictionHistory:history}} as unknown as GooseAdminApi
  const actor = userEvent.setup()
  render(<UsersManagementPage api={api} text={key=>messages[key]} />)
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button',{name:'Edit user'})[0])
  await actor.click(screen.getByRole('tab',{name:'History'}))
  await screen.findByRole('heading',{name:'Restriction history'})
  await waitFor(()=>expect(history).toHaveBeenCalledWith({userId:7,page:1,pageSize:10}))
  expect(screen.queryByLabelText('Private administrator note')).toBeNull()
  expect(screen.queryByText('Private administrator note: Internal only')).toBeNull()
  await actor.click(screen.getByRole('tab',{name:'Account'}))
  expect(screen.getByLabelText<HTMLTextAreaElement>('Restriction reason').value).toBe('Spam')
  const deadline = screen.getByLabelText('Restricted until')
  await actor.click(deadline)
  await actor.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!)
  fireEvent.change(screen.getByLabelText('Time'), { target: { value: '10:00' } })
  await actor.click(screen.getByRole('button',{name:'Done'}))
  await actor.click(screen.getByRole('button',{name:'Save changes'}))
  await waitFor(()=>expect(edit).toHaveBeenCalled())
  expect(edit.mock.calls[0][0]).toMatchObject({userId:7,restrictionStatus:'banned',restrictionReason:'Spam'})
  expect(edit.mock.calls[0][0]).not.toHaveProperty('restrictionNote')
  expect(edit.mock.calls[0][0]).not.toHaveProperty('status')
  expect(edit.mock.calls[0][0].restrictionUntil).toBe(new Date(2026, 9, 6, 10).toISOString())
})
