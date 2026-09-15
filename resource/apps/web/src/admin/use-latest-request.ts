import { useCallback, useEffect, useRef } from "react";

export function useLatestRequest() {
  const version = useRef(0);
  useEffect(
    () => () => {
      version.current++;
    },
    [],
  );
  return useCallback(() => {
    const current = ++version.current;
    return () => version.current === current;
  }, []);
}
