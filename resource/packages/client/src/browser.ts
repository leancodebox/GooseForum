import type { PagePayload } from './contracts/payload.js'
import type { PageClient } from './pages.js'
import { GooseProtocolError } from './http/error.js'

export * from './browser-auth.js'
export * from './document.js'

export function readInitialPayload<TPage extends PagePayload>(pages: PageClient<TPage>, root: ParentNode = document): TPage {
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
