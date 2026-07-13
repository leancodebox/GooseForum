import type { PagePayload } from '@/types/payload'
import { setBaseDocumentTitle } from './document-title'

export function readInitialPayload(): PagePayload {
  const el = document.getElementById('goose-payload')
  if (!el?.textContent) {
    throw new Error('Missing GooseForum payload')
  }
  return JSON.parse(el.textContent) as PagePayload
}

export function updateDocumentMeta(payload: PagePayload) {
  setBaseDocumentTitle(payload.meta.title)
  setMeta('description', payload.meta.description || '')
  setMeta('robots', payload.meta.robots || '')
  setCanonical(payload.meta.canonical || payload.url)
  setLinkRel('prev', payload.meta.prevUrl || '')
  setLinkRel('next', payload.meta.nextUrl || '')
  setPropertyMeta('og:title', payload.meta.openGraph?.title || '')
  setPropertyMeta('og:description', payload.meta.openGraph?.description || '')
  setPropertyMeta('og:type', payload.meta.openGraph?.type || '')
  setPropertyMeta('og:url', payload.meta.openGraph?.url || '')
  setPropertyMeta('og:site_name', payload.meta.openGraph?.siteName || '')
  setPropertyMeta('og:image', payload.meta.openGraph?.image || '')
  setPropertyMeta('article:published_time', payload.meta.openGraph?.publishedTime || '')
  setPropertyMeta('article:modified_time', payload.meta.openGraph?.modifiedTime || '')
  setPropertyMeta('article:author', payload.meta.openGraph?.author || '')
  setPropertyMeta('article:section', payload.meta.openGraph?.section || '')
  setRepeatedPropertyMeta('article:tag', payload.meta.openGraph?.tags || [])
  setMeta('twitter:card', payload.meta.twitter?.card || '')
  setMeta('twitter:title', payload.meta.twitter?.title || '')
  setMeta('twitter:description', payload.meta.twitter?.description || '')
  setMeta('twitter:image', payload.meta.twitter?.image || '')
  setJsonLd(payload.meta.jsonLd)
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
  setLinkRel('canonical', href)
}

function setLinkRel(rel: string, href: string) {
  let el = document.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`)
  if (!href) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('link')
    el.rel = rel
    document.head.appendChild(el)
  }
  el.href = href
}

function setPropertyMeta(property: string, content: string) {
  let el = document.querySelector<HTMLMetaElement>(`meta[property="${property}"]`)
  if (!content) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute('property', property)
    document.head.appendChild(el)
  }
  el.content = content
}

function setRepeatedPropertyMeta(property: string, values: string[]) {
  document.querySelectorAll<HTMLMetaElement>(`meta[property="${property}"]`).forEach((el) => el.remove())
  values.filter(Boolean).forEach((value) => {
    const el = document.createElement('meta')
    el.setAttribute('property', property)
    el.content = value
    document.head.appendChild(el)
  })
}

function setJsonLd(value: unknown) {
  let el = document.querySelector<HTMLScriptElement>('script[data-goose-jsonld="page"]')
  if (!value) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('script')
    el.type = 'application/ld+json'
    el.dataset.gooseJsonld = 'page'
    document.head.appendChild(el)
  }
  el.textContent = JSON.stringify(value)
}
