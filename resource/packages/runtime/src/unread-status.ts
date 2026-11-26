import { useEffect, useState } from "react";
import type { UnreadStatusPayload } from "@gooseforum/client";

export const unreadStatusEvent = "goose:unread-status";

export function announceUnreadStatus(status: Partial<UnreadStatusPayload>) {
  window.dispatchEvent(
    new CustomEvent<Partial<UnreadStatusPayload>>(unreadStatusEvent, {
      detail: status,
    }),
  );
}

export function useUnreadStatus(initial: UnreadStatusPayload) {
  const [unread, setUnread] = useState(initial);
  useEffect(() => {
    setUnread(initial);
    announceUnreadStatus(initial);
  }, [initial]);

  useEffect(() => {
    const updateUnread = (event: Event) => {
      const detail = (event as CustomEvent<Partial<UnreadStatusPayload>>)
        .detail;
      if (!detail) return;
      setUnread((current) => {
        const keys = Object.keys(detail) as (keyof UnreadStatusPayload)[];
        return keys.some(key => current[key] !== detail[key])
          ? { ...current, ...detail }
          : current;
      });
    };
    window.addEventListener(unreadStatusEvent, updateUnread);
    return () => window.removeEventListener(unreadStatusEvent, updateUnread);
  }, []);

  return unread;
}
