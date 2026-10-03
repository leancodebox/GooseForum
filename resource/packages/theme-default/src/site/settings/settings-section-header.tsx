import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

export function SettingsSectionHeader({
  icon: Icon,
  title,
  description,
  actions,
}: {
  icon: LucideIcon;
  title: string;
  description?: string;
  actions?: ReactNode;
}) {
  return (
    <header className="flex items-start justify-between gap-3 px-4 pt-6 pb-1">
      <div className="min-w-0">
        <h2 className="flex items-center gap-2 text-sm font-medium">
          <Icon className="size-3.5 shrink-0 text-muted-foreground" aria-hidden="true" />
          {title}
        </h2>
        {description ? (
          <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {actions}
    </header>
  );
}
