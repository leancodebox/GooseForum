import { createContext, useContext, useLayoutEffect, useState, type ReactNode } from "react";

// Each area supplies its translated controls and navigation context. The host
// reconciles that tree in one position even when the area below it is replaced.
const PublishHeader = createContext<((header: ReactNode) => void) | null>(null);

export function HeaderHost({ children }: { children: ReactNode }) {
  const [header, publish] = useState<ReactNode>(null);
  return <PublishHeader value={publish}>{header}{children}</PublishHeader>;
}

export function HeaderSlot({ children }: { children: ReactNode }) {
  const publish = useContext(PublishHeader);
  useLayoutEffect(() => {
    if (!publish) return;
    publish(children);
    return () => publish(null);
  }, [publish, children]);
  return publish ? null : children;
}
