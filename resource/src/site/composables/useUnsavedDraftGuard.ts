import { onBeforeUnmount, onMounted, ref, type ComputedRef } from 'vue'
import { useRouter, type RouteLocationNormalized } from 'vue-router'

interface UnsavedDraftGuardOptions {
  hasUnsavedChanges: ComputedRef<boolean>
  canSaveDraft: ComputedRef<boolean>
  saveDraftBeforeLeave: () => Promise<boolean>
}

export function useUnsavedDraftGuard(options: UnsavedDraftGuardOptions) {
  const router = useRouter()
  const leavePromptOpen = ref(false)
  const forcedNavigation = ref(false)
  let removeRouteGuard: (() => void) | undefined
  let resolveLeavePrompt: ((allow: boolean) => void) | undefined

  onMounted(() => {
    removeRouteGuard = router.beforeEach((to) => confirmRouteLeave(to))
    window.addEventListener('beforeunload', handleBeforeUnload)
  })

  onBeforeUnmount(() => {
    removeRouteGuard?.()
    resolveLeavePrompt?.(true)
    window.removeEventListener('beforeunload', handleBeforeUnload)
  })

  function forceNextNavigation() {
    forcedNavigation.value = true
  }

  function handleBeforeUnload(event: BeforeUnloadEvent) {
    if (forcedNavigation.value || !options.hasUnsavedChanges.value) return
    event.preventDefault()
    event.returnValue = ''
  }

  function confirmRouteLeave(_to: RouteLocationNormalized) {
    if (forcedNavigation.value || !options.hasUnsavedChanges.value) return true
    leavePromptOpen.value = true
    return new Promise<boolean>((resolve) => {
      resolveLeavePrompt = resolve
    })
  }

  function closeLeavePrompt() {
    leavePromptOpen.value = false
    resolveLeavePrompt?.(false)
    resolveLeavePrompt = undefined
  }

  function discardAndLeave() {
    forceNextNavigation()
    leavePromptOpen.value = false
    resolveLeavePrompt?.(true)
    resolveLeavePrompt = undefined
  }

  async function saveDraftAndLeave() {
    if (!options.canSaveDraft.value) return
    const saved = await options.saveDraftBeforeLeave()
    if (!saved) return
    leavePromptOpen.value = false
    resolveLeavePrompt?.(true)
    resolveLeavePrompt = undefined
  }

  return {
    leavePromptOpen,
    forceNextNavigation,
    closeLeavePrompt,
    discardAndLeave,
    saveDraftAndLeave,
  }
}
