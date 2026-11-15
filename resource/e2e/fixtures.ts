import { test as base, expect } from '@playwright/test'
import { fileURLToPath } from 'node:url'

// These UI tests run Vite without Go. Individual tests override routes as needed.
export const test = base.extend({
  page: async ({ page }, use) => {
    await page.route('**/api/forum/status', route => route.fulfill({
      json: { code: 0, result: { deploymentId: 'e2e', unread: { notifications: false, messages: false } } },
    }))
    await page.route('**/file/attachments/policy', route => route.fulfill({
      json: { code: 0, result: { enabled: true, maxFileSizeKb: 5120, maxFilesPerPost: 5, extensions: ['.txt', '.pdf', '.zip'] } },
    }))
    await page.route('**/static/**', route => route.fulfill({
      path: fileURLToPath(new URL(`../${new URL(route.request().url()).pathname.slice(1)}`, import.meta.url)),
    }))
    await use(page)
  },
})
export { expect }
