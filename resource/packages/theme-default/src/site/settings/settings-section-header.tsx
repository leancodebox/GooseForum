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
    <header className="px-4 pt-6 pb-1">
      <div className="flex min-h-8 items-center justify-between gap-3">
        <h2 className="flex min-w-0 items-center gap-2 text-sm font-medium">
          <Icon className="size-3.5 shrink-0 text-muted-foreground" aria-hidden="true" />
          {title}
        </h2>
        {actions ? <div className="flex shrink-0 items-center">{actions}</div> : null}
      </div>
      {description ? (
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{description}</p>
      ) : null}
    </header>
  );
}
