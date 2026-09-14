import { vi } from 'vitest'

vi.stubGlobal('ResizeObserver', class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
})
