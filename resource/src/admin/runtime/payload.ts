import type { AdminPayload } from '@/admin/types'

export function readAdminPayload(): AdminPayload {
  const el = document.getElementById('goose-payload')
  if (!el?.textContent) {
    throw new Error('Missing GooseForum admin payload')
  }
  return JSON.parse(el.textContent) as AdminPayload
}
