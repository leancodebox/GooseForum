export interface UserCardTarget {
  id: number
  username: string
  avatarUrl: string
}

export interface UserCardShowDetail {
  user: UserCardTarget
  target: HTMLElement
}

export function showUserCard(user: UserCardTarget, event: MouseEvent | FocusEvent) {
  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return

  window.dispatchEvent(
    new CustomEvent<UserCardShowDetail>('goose:user-card-show', {
      detail: { user, target },
    }),
  )
}

export function scheduleHideUserCard() {
  window.dispatchEvent(new CustomEvent('goose:user-card-hide'))
}
