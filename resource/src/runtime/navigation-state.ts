import { readonly, ref } from 'vue'

const navigating = ref(false)

export function useNavigationState() {
  return {
    navigating: readonly(navigating),
    setNavigating(value: boolean) {
      navigating.value = value
    },
  }
}
