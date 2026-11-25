import type { ReactNode } from "react";
import { DynamicIcon, dynamicIconImports, iconNames } from "lucide-react/dynamic";

export { iconNames as navigationIconNames };

export default function DynamicNavigationIcon({ name, fallback }: { name: string; fallback?: ReactNode }) {
  if (!Object.hasOwn(dynamicIconImports, name)) return fallback ?? null;

  return (
    <DynamicIcon
      name={name as keyof typeof dynamicIconImports}
      aria-hidden="true"
      className="size-4 shrink-0"
      fallback={() => <>{fallback}</>}
    />
  );
}
