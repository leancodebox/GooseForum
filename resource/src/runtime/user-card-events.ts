export interface UserCardTarget {
  id: number
  username: string
  avatarUrl: string
}

export interface UserCardShowDetail {
  user: UserCardTarget
  target: HTMLElement
}

export function showUserCard(user: UserCardTarget, event: MouseEvent) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

  event.preventDefault()
  event.stopPropagation()

  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return

  window.dispatchEvent(
    new CustomEvent<UserCardShowDetail>('goose:user-card-show', {
      detail: { user, target },
    }),
  )
}
