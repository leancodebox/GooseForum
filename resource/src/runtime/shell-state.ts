import { reactive } from 'vue'

const shellState = reactive({
  headerTitle: '',
  showHeaderTitle: false,
})

export function useShellState() {
  return shellState
}

export function resetShellState() {
  shellState.headerTitle = ''
  shellState.showHeaderTitle = false
}
