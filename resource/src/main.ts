import { createApp, h, shallowRef } from 'vue'
import App from './App.vue'
import './styles/resource.css'
import { readInitialPayload, updateDocumentMeta } from './runtime/payload'
import { installNavigation } from './runtime/router'
import type { PagePayload } from './types/payload'

const payload = shallowRef<PagePayload>(readInitialPayload())

history.replaceState({ goose: true, payload: payload.value }, '', window.location.href)

createApp({
  setup() {
    return () => h(App, { payload: payload.value })
  },
}).mount('#goose-app')

installNavigation((nextPayload) => {
  payload.value = nextPayload
  updateDocumentMeta(nextPayload)
})

window.addEventListener('goose:page', (event) => {
  const nextPayload = (event as CustomEvent<PagePayload>).detail
  if (!nextPayload) return
  payload.value = nextPayload
  updateDocumentMeta(nextPayload)
})
