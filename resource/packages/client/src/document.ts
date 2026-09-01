import type { PagePayload } from './contracts/payload.js'

/** Keep browser metadata in sync after client-side page navigation. */
export function updateDocumentMetadata(payload: PagePayload, document: Document = globalThis.document) {
  document.title = payload.meta.title
  setNamedMeta(document, 'description', payload.meta.description || '')
  setNamedMeta(document, 'robots', payload.meta.robots || '')
  setLink(document, 'canonical', payload.meta.canonical || payload.url)
  setLink(document, 'prev', payload.meta.prevUrl || '')
  setLink(document, 'next', payload.meta.nextUrl || '')
  setPropertyMeta(document, 'og:title', payload.meta.openGraph?.title || '')
  setPropertyMeta(document, 'og:description', payload.meta.openGraph?.description || '')
  setPropertyMeta(document, 'og:type', payload.meta.openGraph?.type || '')
  setPropertyMeta(document, 'og:url', payload.meta.openGraph?.url || '')
  setPropertyMeta(document, 'og:site_name', payload.meta.openGraph?.siteName || '')
  setPropertyMeta(document, 'og:image', payload.meta.openGraph?.image || '')
  setPropertyMeta(document, 'article:published_time', payload.meta.openGraph?.publishedTime || '')
  setPropertyMeta(document, 'article:modified_time', payload.meta.openGraph?.modifiedTime || '')
  setPropertyMeta(document, 'article:author', payload.meta.openGraph?.author || '')
  setPropertyMeta(document, 'article:section', payload.meta.openGraph?.section || '')
  setRepeatedPropertyMeta(document, 'article:tag', payload.meta.openGraph?.tags || [])
  setNamedMeta(document, 'twitter:card', payload.meta.twitter?.card || '')
  setNamedMeta(document, 'twitter:title', payload.meta.twitter?.title || '')
  setNamedMeta(document, 'twitter:description', payload.meta.twitter?.description || '')
  setNamedMeta(document, 'twitter:image', payload.meta.twitter?.image || '')
  setJsonLd(document, payload.meta.jsonLd)
}

function setNamedMeta(document: Document, name: string, content: string) {
  let element = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!content) return element?.remove()
  if (!element) {
    element = document.createElement('meta')
    element.name = name
    document.head.appendChild(element)
  }
  element.content = content
}

function setPropertyMeta(document: Document, property: string, content: string) {
  let element = document.querySelector<HTMLMetaElement>(`meta[property="${property}"]`)
  if (!content) return element?.remove()
  if (!element) {
    element = document.createElement('meta')
    element.setAttribute('property', property)
    document.head.appendChild(element)
  }
  element.content = content
}

function setLink(document: Document, rel: string, href: string) {
  let element = document.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`)
  if (!href) return element?.remove()
  if (!element) {
    element = document.createElement('link')
    element.rel = rel
    document.head.appendChild(element)
  }
  element.href = href
}

function setRepeatedPropertyMeta(document: Document, property: string, values: string[]) {
  document.querySelectorAll(`meta[property="${property}"]`).forEach((element) => element.remove())
  for (const value of values.filter(Boolean)) {
    const element = document.createElement('meta')
    element.setAttribute('property', property)
    element.content = value
    document.head.appendChild(element)
  }
}

function setJsonLd(document: Document, value: unknown) {
  let element = document.querySelector<HTMLScriptElement>('script[data-goose-jsonld="page"]')
  if (!value) return element?.remove()
  if (!element) {
    element = document.createElement('script')
    element.type = 'application/ld+json'
    element.dataset.gooseJsonld = 'page'
    document.head.appendChild(element)
  }
  element.textContent = JSON.stringify(value)
}
