import { createApp, h, shallowRef } from 'vue'
import App from './App.vue'
import './styles/resource.css'
import { readInitialPayload, updateDocumentMeta } from './runtime/payload'
import { installNavigation, preparePayload } from './runtime/router'
import { resetShellState } from './runtime/shell-state'

const initialPayload = readInitialPayload()
const initialPage = await preparePayload(initialPayload)
const currentPage = shallowRef(initialPage)

history.replaceState({ goose: true, payload: currentPage.value.payload }, '', window.location.href)

createApp({
  setup() {
    return () => h(App, {
      payload: currentPage.value.payload,
      component: currentPage.value.component,
    })
  },
}).mount('#goose-app')

function commitPage(nextPage: typeof initialPage) {
  resetShellState()
  currentPage.value = nextPage
  updateDocumentMeta(nextPage.payload)
}

installNavigation((nextPage) => {
  commitPage(nextPage)
})

window.addEventListener('goose:page', async (event) => {
  const nextPayload = event instanceof CustomEvent ? event.detail : undefined
  if (!nextPayload) return
  commitPage(await preparePayload(nextPayload))
})
