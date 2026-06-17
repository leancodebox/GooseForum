import { createApp, h, shallowRef } from 'vue'
import App from '@/site/App.vue'
import '@/styles/resource.css'
import { readInitialPayload, updateDocumentMeta } from '@/runtime/payload'
import { installNavigation, preparePayload } from '@/runtime/router'
import { currentLocale, i18n } from '@/runtime/i18n'
import { hydrateFlashMessages } from '@/runtime/flash-message'
import { applySiteThemePayload, applyStoredTheme } from '@/runtime/site-theme'
import PayloadRouteView from '@/site/components/PayloadRouteView.vue'

const initialPayload = readInitialPayload()
const initialPage = await preparePayload(initialPayload)
const currentPage = shallowRef(initialPage)

if (typeof window !== 'undefined' && 'scrollRestoration' in window.history) {
  window.history.scrollRestoration = 'manual'
}

document.documentElement.lang = currentLocale()
applySiteThemePayload(initialPayload.layout.theme)
applyStoredTheme()

function commitPage(nextPage: typeof initialPage) {
  currentPage.value = nextPage
  applySiteThemePayload(nextPage.payload.layout.theme)
  updateDocumentMeta(nextPage.payload)
}

const router = installNavigation(initialPage, PayloadRouteView, (nextPage) => {
  commitPage(nextPage)
})

const app = createApp({
  setup() {
    return () => h(App, {
      page: currentPage.value,
    })
  },
})

app.use(i18n)
app.use(router)
await router.isReady()
app.mount('#goose-app')
hydrateFlashMessages()

window.addEventListener('goose:page', async (event) => {
  const nextPayload = event instanceof CustomEvent ? event.detail : undefined
  if (!nextPayload) return
  commitPage(await preparePayload(nextPayload))
})
