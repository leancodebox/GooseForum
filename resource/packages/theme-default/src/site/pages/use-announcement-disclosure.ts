import { useState } from "react";

const storageKey = "goose:announcement:disclosure";

function restore(version: string) {
  let open = true;
  try {
    const saved: unknown = JSON.parse(localStorage.getItem(storageKey) || "null");
    if (saved && typeof saved === "object" && "version" in saved &&
      saved.version === version && "open" in saved && typeof saved.open === "boolean") {
      open = saved.open;
    }
  } catch {}
  return { version, open, animate: false };
}

export function useAnnouncementDisclosure(version: string) {
  const [state, setState] = useState(() => restore(version));
  // Replace the old version before children commit, without animating restoration.
  if (state.version !== version) setState(restore(version));

  function setOpen(open: boolean) {
    setState({ version, open, animate: true });
    try {
      localStorage.setItem(storageKey, JSON.stringify({ version, open }));
    } catch {}
  }

  function finishAnimation() {
    setState(previous => previous.animate ? { ...previous, animate: false } : previous);
  }

  return { open: state.open, animate: state.animate, setOpen, finishAnimation };
}
