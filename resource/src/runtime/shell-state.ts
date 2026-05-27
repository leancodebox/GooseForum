import { reactive } from 'vue'

export interface ShellHeaderTag {
  id: number | string
  name: string
  color?: string
}

const shellState = reactive({
  headerTitle: '',
  headerTags: [] as ShellHeaderTag[],
  showHeaderTitle: false,
})

export function useShellState() {
  return shellState
}

export function resetShellState() {
  shellState.headerTitle = ''
  shellState.headerTags = []
  shellState.showHeaderTitle = false
}
