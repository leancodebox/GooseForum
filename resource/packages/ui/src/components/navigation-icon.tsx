import { lazy, Suspense, type ReactNode } from "react";

const DynamicNavigationIcon = lazy(() => import("./navigation-icon-dynamic"));

export function NavigationIcon({ name, fallback }: { name?: string; fallback?: ReactNode }) {
  if (!name) return fallback ?? null;

  return (
    <Suspense fallback={fallback}>
      <DynamicNavigationIcon key={name} name={name} fallback={fallback} />
    </Suspense>
  );
}
