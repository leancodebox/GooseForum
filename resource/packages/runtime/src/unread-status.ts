import type { UnreadStatusPayload } from "@gooseforum/client";

export const unreadStatusEvent = "goose:unread-status";

export function announceUnreadStatus(status: Partial<UnreadStatusPayload>) {
  window.dispatchEvent(
    new CustomEvent<Partial<UnreadStatusPayload>>(unreadStatusEvent, {
      detail: status,
    }),
  );
}
