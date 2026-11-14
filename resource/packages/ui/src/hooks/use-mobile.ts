import * as React from "react"

const MOBILE_BREAKPOINT = 768

export function useIsMobile() {
  const query = `(max-width: ${MOBILE_BREAKPOINT - 1}px)`
  const subscribe = React.useCallback((onStoreChange: () => void) => {
    const mql = window.matchMedia(query)
    mql.addEventListener("change", onStoreChange)
    return () => mql.removeEventListener("change", onStoreChange)
  }, [query])
  const getSnapshot = React.useCallback(
    () => window.innerWidth < MOBILE_BREAKPOINT,
    [],
  )
  const getServerSnapshot = React.useCallback(() => false, [])

  return React.useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
}
