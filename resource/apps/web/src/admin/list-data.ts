import type { GooseAdminApi } from '@gooseforum/client'
import { useCallback, useEffect, useState } from 'react'
import { normalizeAdminPath } from './navigation'

export interface UsersQuery { page: number; pageSize: number; search: string }
export interface PostsQuery extends UsersQuery { kind: 'topic' | 'post'; moderationStatus: string; categoryId: string }
export interface UsersPageData { query: UsersQuery; result: Awaited<ReturnType<GooseAdminApi['users']['list']>> }
export interface PostsPageData { query: PostsQuery; result: Awaited<ReturnType<GooseAdminApi['topics']['list']>> | Awaited<ReturnType<GooseAdminApi['topics']['reviewPosts']>>; categories: Awaited<ReturnType<GooseAdminApi['categories']['list']>> }
export type AdminListData = { kind: 'users'; data: UsersPageData } | { kind: 'posts'; data: PostsPageData }

function positiveInteger(value: string | null, fallback: number) {
  const number = Number(value)
  return Number.isSafeInteger(number) && number > 0 ? number : fallback
}

export function usersQuery(search: string): UsersQuery {
  const params = new URLSearchParams(search)
  const size = positiveInteger(params.get('pageSize'), 10)
  return { page: positiveInteger(params.get('page'), 1), pageSize: [10, 20, 30, 50].includes(size) ? size : 10, search: params.get('search') || '' }
}

export function postsQuery(search: string): PostsQuery {
  const params = new URLSearchParams(search)
  const query = usersQuery(search)
  const status = params.get('moderationStatus') || 'all'
  return { ...query, pageSize: [10, 20, 50].includes(query.pageSize) ? query.pageSize : 10, kind: params.get('kind') === 'post' ? 'post' : 'topic', moderationStatus: ['all', 'pending', 'approved', 'rejected', 'none'].includes(status) ? status : 'all', categoryId: String(positiveInteger(params.get('categoryId'), 0) || 'all') }
}

export function loadUsers(api: GooseAdminApi, query: UsersQuery) {
  return api.users.list({ page: query.page, pageSize: query.pageSize, username: query.search || undefined })
}

export function loadPosts(api: GooseAdminApi, query: PostsQuery) {
  const input = { page: query.page, pageSize: query.pageSize, search: query.search || undefined, moderationStatus: query.moderationStatus === 'all' ? undefined : query.moderationStatus }
  return query.kind === 'post' ? api.topics.reviewPosts(input) : api.topics.list({ ...input, categoryId: query.categoryId === 'all' ? undefined : Number(query.categoryId) })
}

export async function prepareAdminListData(api: GooseAdminApi, url: URL): Promise<AdminListData | undefined> {
  const path = normalizeAdminPath(url.pathname)
  if (path === '/admin/users') {
    const query = usersQuery(url.search)
    return { kind: 'users', data: { query, result: await loadUsers(api, query) } }
  }
  if (path === '/admin/posts') {
    const query = postsQuery(url.search)
    const [result, categories] = await Promise.all([loadPosts(api, query), api.categories.list().catch(() => [])])
    return { kind: 'posts', data: { query, result, categories } }
  }
}

export function useAdminListQuery<Query extends UsersQuery>(path: string, initial: Query, onUrlChange?: (url: string) => void) {
  const [query, setQuery] = useState(initial)
  useEffect(() => {
    if (normalizeAdminPath(window.location.pathname) !== path) return
    const url = new URL(window.location.href)
    url.pathname = path
    const defaults: Record<string, string | number> = { page: 1, pageSize: 10, search: '', kind: 'topic', moderationStatus: 'all', categoryId: 'all' }
    for (const [key, value] of Object.entries(query)) {
      if (value === defaults[key]) url.searchParams.delete(key)
      else url.searchParams.set(key, String(value))
    }
    window.history.replaceState(window.history.state, '', url)
    onUrlChange?.(url.pathname + url.search)
  }, [path, query, onUrlChange])
  const updateQuery = useCallback((patch: Partial<Query>) => {
    setQuery(current => ({ ...current, page: 1, ...patch }))
  }, [])
  return [query, updateQuery] as const
}
