import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import type { GooseSiteApi } from '@gooseforum/client'
import { SiteStatusNotice } from './site-status'

beforeEach(() => {
  vi.useFakeTimers()
  document.documentElement.lang = 'en'
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' })
})
afterEach(() => { cleanup(); vi.useRealTimers(); vi.restoreAllMocks() })
const apiFor = (status: ReturnType<typeof vi.fn>) => ({ status }) as unknown as GooseSiteApi
const flush = async () => { await act(async () => { await Promise.resolve() }) }

it('uses initial HTML baseline, emits unread, and suppresses the same dismissed version', async () => {
  const status = vi.fn().mockResolvedValue({ deploymentId: 'new', unread: { notifications: true } })
  const unread = vi.fn()
  window.addEventListener('goose:unread-status', unread)
  render(<SiteStatusNotice api={apiFor(status)} initialDeployment="old" />)
  await flush()
  expect(screen.getByRole('status').textContent).toContain('site has been updated')
  expect(unread).toHaveBeenCalledTimes(1)
  fireEvent.click(screen.getByText('Later'))
  await act(async () => { await vi.advanceTimersByTimeAsync(30_000) })
  expect(screen.queryByRole('status')).toBeNull()
  status.mockResolvedValue({ deploymentId: 'newer' })
  await act(async () => { await vi.advanceTimersByTimeAsync(30_000) })
  expect(screen.getByRole('status')).toBeTruthy()
  window.removeEventListener('goose:unread-status', unread)
})

it('pauses hidden tabs, checks on return, and never overlaps requests', async () => {
  let resolve!: (value: unknown) => void
  const status = vi.fn((_signal?: AbortSignal) => new Promise(r => { resolve = r }))
  const { unmount } = render(<SiteStatusNotice api={apiFor(status)} initialDeployment="old" />)
  await flush()
  fireEvent(document, new Event('visibilitychange'))
  expect(status).toHaveBeenCalledTimes(1)
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'hidden' })
  await act(async () => { resolve({ deploymentId: 'old' }) })
  await act(async () => { await vi.advanceTimersByTimeAsync(30_000) })
  expect(status).toHaveBeenCalledTimes(1)
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' })
  fireEvent(document, new Event('visibilitychange'))
  expect(status).toHaveBeenCalledTimes(2)
  const signal = status.mock.calls[1][0] as AbortSignal
  unmount()
  expect(signal.aborted).toBe(true)
})

it('does not mistake network failure for an update; resource errors offer recovery', async () => {
  render(<SiteStatusNotice api={apiFor(vi.fn().mockRejectedValue(new Error('offline')))} initialDeployment="old" admin />)
  await flush()
  expect(screen.queryByRole('status')).toBeNull()
  fireEvent(window, new Event('vite:preloadError'))
  expect(screen.getByRole('status').textContent).toContain('resources failed to load')
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  fireEvent.click(screen.getByText('Refresh page'))
  expect(confirm).toHaveBeenCalledOnce()
})

it('does not overwrite a newer read action with an old poll response', async () => {
  let resolve!: (value: unknown) => void
  const status = vi.fn(() => new Promise(r => { resolve = r }))
  const unread = vi.fn()
  window.addEventListener('goose:unread-status', unread)
  render(<SiteStatusNotice api={apiFor(status)} initialDeployment="old" />)
  await flush()
  fireEvent(window, new CustomEvent('goose:unread-status', { detail: { notifications: false } }))
  await act(async () => { resolve({ deploymentId: 'new', unread: { notifications: true } }) })
  expect(unread).toHaveBeenCalledTimes(1)
  // Discard only stale unread data; deployment detection still works.
  expect(screen.getByRole('status').textContent).toContain('site has been updated')
  window.removeEventListener('goose:unread-status', unread)
})
