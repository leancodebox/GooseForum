/** A theme transition owns only its snapshots; page/editor state stays mounted. */
export function createThemeTransition() {
  let active: ViewTransition | undefined;
  let revision = 0;
  const className = "goose-theme-transition";

  function cancel() {
    revision++;
    active?.skipTransition();
    active = undefined;
    document.documentElement.classList.remove(className);
  }

  function apply(update: () => void, animate = true) {
    cancel();
    const request = revision;
    if (
      !animate ||
      typeof document.startViewTransition !== "function" ||
      document.visibilityState === "hidden" ||
      window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
    ) {
      update();
      return;
    }

    let applied = false;
    const updateOnce = () => {
      if (request !== revision || applied) return;
      applied = true;
      update();
    };
    document.documentElement.classList.add(className);
    try {
      const transition = document.startViewTransition(updateOnce);
      active = transition;
      // Snapshot failure must not prevent the requested theme from being applied.
      void transition.updateCallbackDone.catch(updateOnce);
      void transition.finished.catch(() => {}).finally(() => {
        if (request !== revision) return;
        active = undefined;
        document.documentElement.classList.remove(className);
      });
    } catch {
      document.documentElement.classList.remove(className);
      updateOnce();
    }
  }

  return { apply, cancel };
}
