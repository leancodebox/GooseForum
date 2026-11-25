import { createContext, Suspense, use, type ReactNode } from "react";
import { Icon, type LucideIconData } from "lucide-react";

type NavigationIconProps = { name?: string; fallback?: ReactNode };

export const PageNavigationIcons = createContext<Record<string, LucideIconData> | undefined>(undefined);

type IconCatalog = Record<string, LucideIconData>;
let readyCatalog: IconCatalog | null | undefined;
let catalogRequest: Promise<IconCatalog | null> | undefined;

export function loadNavigationIconCatalog() {
  if (readyCatalog) return Promise.resolve(readyCatalog);
  return catalogRequest ??= import("./navigation-icon-dynamic")
    .then(module => module.fetchCatalog())
    .catch(() => null)
    .then(icons => {
      readyCatalog = icons;
      catalogRequest = undefined;
      return icons;
    });
}

export function preloadNavigationIcons(names: Iterable<string | undefined>) {
  for (const name of names) {
    if (name) return loadNavigationIconCatalog();
  }
}

export function NavigationIconContent({ name, fallback }: NavigationIconProps) {
  const pageIcons = use(PageNavigationIcons);
  if (!name) return fallback ?? null;
  const icons = pageIcons ?? (readyCatalog === undefined ? use(loadNavigationIconCatalog()) : readyCatalog);
  const icon = icons && Object.hasOwn(icons, name) ? icons[name] : null;
  return icon ? <Icon icon={icon} aria-hidden="true" className="size-4 shrink-0" /> : fallback ?? null;
}

export function NavigationIcon(props: NavigationIconProps) {
  return (
    <Suspense fallback={<span aria-hidden="true" className="size-4 shrink-0" />}>
      <NavigationIconContent {...props} />
    </Suspense>
  );
}
