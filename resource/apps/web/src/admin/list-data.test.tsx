import { StrictMode } from 'react'
import type { GooseAdminApi } from '@gooseforum/client'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { prepareAdminListData, postsQuery, usersQuery } from './list-data'
import { UsersManagementPage } from './pages/users-management-page'

beforeEach(() => { window.history.replaceState(null, '', '/admin/users') })
afterEach(cleanup)

describe('admin list preparation', () => {
  it('validates pagination from URLs and loads the selected reply filters', async () => {
    expect(usersQuery('?page=-1&pageSize=900')).toEqual({ page: 1, pageSize: 10, search: '' })
    const reviewPosts = vi.fn().mockResolvedValue({ list: [], total: 0, page: 3 })
    const categories = vi.fn().mockResolvedValue([])
    const api = { topics: { reviewPosts }, categories: { list: categories } } as unknown as GooseAdminApi
    const url = new URL('http://localhost/admin/posts?kind=post&page=3&moderationStatus=pending&search=hello')
    const data = await prepareAdminListData(api, url)
    expect(data?.kind).toBe('posts')
    expect(reviewPosts).toHaveBeenCalledWith({ page: 3, pageSize: 10, moderationStatus: 'pending', search: 'hello' })
    expect(postsQuery('?categoryId=oops&kind=unknown').categoryId).toBe('all')
  })

  it('renders prepared data without a second fetch, then fetches and records a new search', async () => {
    window.history.replaceState(null, '', '/admin/users/')
    const list = vi.fn().mockResolvedValue({ list: [], total: 0, page: 1 })
    const api = { users: { list } } as unknown as GooseAdminApi
    const initialData = { query: usersQuery(''), result: { list: [], total: 0, page: 1 } }
    render(<StrictMode><UsersManagementPage api={api} text={key => key} initialData={initialData} /></StrictMode>)
    expect(list).not.toHaveBeenCalled()
    expect(window.location.pathname).toBe('/admin/users')
    fireEvent.change(screen.getByPlaceholderText('search'), { target: { value: 'alice' } })
    fireEvent.click(screen.getByRole('button', { name: 'searchAction' }))
    await waitFor(() => expect(list).toHaveBeenCalledTimes(1))
    expect(list).toHaveBeenCalledWith({ page: 1, pageSize: 10, username: 'alice' })
    expect(new URLSearchParams(window.location.search).get('search')).toBe('alice')
    fireEvent.click(screen.getByRole('button', { name: 'refresh' }))
    await waitFor(() => expect(list).toHaveBeenCalledTimes(2))
  })
})
