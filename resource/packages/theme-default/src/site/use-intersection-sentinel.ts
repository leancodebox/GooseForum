import { useEffect, useRef, type RefObject } from "react";

export function useIntersectionSentinel(
  sentinel: RefObject<Element | null>,
  onIntersect: () => void,
  enabled = true,
  rootMargin = "480px 0px",
) {
  const onIntersectRef = useRef(onIntersect);

  useEffect(() => {
    onIntersectRef.current = onIntersect;
  }, [onIntersect]);

  useEffect(() => {
    if (!enabled || !sentinel.current || !("IntersectionObserver" in window)) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onIntersectRef.current();
      },
      { rootMargin },
    );
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [enabled, rootMargin, sentinel]);
}
