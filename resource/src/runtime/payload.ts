import type { PagePayload } from '@/types/payload'

export function readInitialPayload(): PagePayload {
  const el = document.getElementById('goose-payload')
  if (!el?.textContent) {
    throw new Error('Missing GooseForum payload')
  }
  return JSON.parse(el.textContent) as PagePayload
}

export function updateDocumentMeta(payload: PagePayload) {
  document.title = payload.meta.title
  setMeta('description', payload.meta.description || '')
  setMeta('robots', payload.meta.robots || '')
  setCanonical(payload.meta.canonical || payload.url)
}

function setMeta(name: string, content: string) {
  let el = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!content) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('meta')
    el.name = name
    document.head.appendChild(el)
  }
  el.content = content
}

function setCanonical(href: string) {
  let el = document.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!href) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('link')
    el.rel = 'canonical'
    document.head.appendChild(el)
  }
  el.href = href
}
