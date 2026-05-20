const unreadMessageAlerts = ['【新私信】', '【请查看】'] as const
const blinkIntervalMs = 1200

let baseTitle = document.title
let hasUnreadMessages = false
let unreadAlertIndex = 0
let blinkTimer: number | undefined

function applyDocumentTitle() {
  document.title = hasUnreadMessages ? `${unreadMessageAlerts[unreadAlertIndex]} ${baseTitle}` : baseTitle
}

function stopUnreadTitleBlink() {
  if (blinkTimer !== undefined) {
    window.clearInterval(blinkTimer)
    blinkTimer = undefined
  }
  unreadAlertIndex = 0
}

function startUnreadTitleBlink() {
  if (blinkTimer !== undefined) return
  unreadAlertIndex = 0
  blinkTimer = window.setInterval(() => {
    unreadAlertIndex = (unreadAlertIndex + 1) % unreadMessageAlerts.length
    applyDocumentTitle()
  }, blinkIntervalMs)
}

export function setBaseDocumentTitle(title: string) {
  baseTitle = title || document.title
  applyDocumentTitle()
}

export function setUnreadMessagesDocumentTitle(hasUnread: boolean) {
  hasUnreadMessages = hasUnread
  if (hasUnreadMessages) {
    startUnreadTitleBlink()
  } else {
    stopUnreadTitleBlink()
  }
  applyDocumentTitle()
}
