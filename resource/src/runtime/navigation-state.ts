import { readonly, ref } from 'vue'

const navigating = ref(false)
const SHOW_DELAY_MS = 120
const MIN_VISIBLE_MS = 280

let pending = false
let visibleSince = 0
let showTimer: number | undefined
let hideTimer: number | undefined

function clearTimer(timer: number | undefined) {
  if (timer !== undefined) {
    window.clearTimeout(timer)
  }
}

function show() {
  showTimer = undefined
  if (!pending || navigating.value) return
  visibleSince = Date.now()
  navigating.value = true
}

function hide() {
  hideTimer = undefined
  if (pending || !navigating.value) return
  navigating.value = false
}

export function useNavigationState() {
  return {
    navigating: readonly(navigating),
    setNavigating(value: boolean) {
      pending = value

      if (value) {
        clearTimer(hideTimer)
        hideTimer = undefined
        if (!navigating.value && showTimer === undefined) {
          showTimer = window.setTimeout(show, SHOW_DELAY_MS)
        }
        return
      }

      clearTimer(showTimer)
      showTimer = undefined
      if (!navigating.value) return

      const elapsed = Date.now() - visibleSince
      const remaining = Math.max(MIN_VISIBLE_MS - elapsed, 0)
      clearTimer(hideTimer)
      hideTimer = window.setTimeout(hide, remaining)
    },
  }
}
