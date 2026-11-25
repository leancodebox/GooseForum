import { NavigationIcon, NavigationIconContent } from "./navigation-icon";
import { cn } from "../lib/utils";

const iconPrefix = "lucide:";

export function categoryIconName(icon?: string) {
  return icon?.startsWith(iconPrefix) ? icon.slice(iconPrefix.length) : undefined;
}

export function encodeCategoryIcon(name: string) {
  return name ? iconPrefix + name : "";
}

export function CategoryIcon({ icon, color, className, suspend }: {
  icon?: string;
  color?: string;
  className?: string;
  suspend?: boolean;
}) {
  if (!icon) return null;
  const name = categoryIconName(icon);
  const LucideIcon = suspend ? NavigationIconContent : NavigationIcon;

  return (
    <span
      className={cn("inline-flex shrink-0 items-center justify-center text-base leading-none", className)}
      style={{ color }}
      aria-hidden="true"
    >
      {name !== undefined ? (
        <LucideIcon name={name} fallback={<span className="size-4 rounded-sm bg-current/25" />} />
      ) : /^(https?:\/\/|\/)/.test(icon) ? (
        <img src={icon} alt="" loading="lazy" decoding="async" className="size-5 rounded object-cover" />
      ) : icon}
    </span>
  );
}

export function CategoryNavIcon({ icon, color, suspend }: { icon?: string; color?: string; suspend?: boolean }) {
  if (icon) {
    return <CategoryIcon icon={icon} color={color || undefined} className="size-4 overflow-hidden [&_img]:size-4" suspend={suspend} />;
  }

  return (
    <span className="grid size-4 shrink-0 place-items-center" aria-hidden="true">
      <span className="size-2 rounded-full" style={{ backgroundColor: color }} />
    </span>
  );
}
