import type { AnyPagePayload } from './contracts/page-components.js'
import type { PageClient } from './pages.js'
import { GooseProtocolError } from './http/error.js'

export function readInitialPayload(pages: PageClient, root: ParentNode = document): AnyPagePayload {
  const element = root.querySelector('#goose-payload')
  if (!element?.textContent) {
    throw new GooseProtocolError('Missing GooseForum page payload')
  }
  try {
    return pages.parse(JSON.parse(element.textContent))
  } catch (error) {
    if (error instanceof GooseProtocolError) throw error
    throw new GooseProtocolError('GooseForum page payload contains invalid JSON', { cause: error })
  }
}
