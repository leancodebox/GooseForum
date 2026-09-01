import { createGooseClient, updateDocumentMetadata, type PagePayload } from '@gooseforum/client'
import { setBaseDocumentTitle } from './document-title'

const client = createGooseClient()

export function readInitialPayload(): PagePayload {
  return client.readInitialPayload()
}

export function updateDocumentMeta(payload: PagePayload) {
  updateDocumentMetadata(payload)
  // Preserve the site's unread-message title behavior on top of SDK metadata.
  setBaseDocumentTitle(payload.meta.title)
}
