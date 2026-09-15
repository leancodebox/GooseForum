import type { ReactNode } from "react";
import { cn } from "@gooseforum/ui/lib/utils";

export function PageHeader({
  title,
  description,
  badge,
  actions,
  compact = false,
  divided = true,
}: {
  title: string;
  description?: string;
  badge?: ReactNode;
  actions?: ReactNode;
  compact?: boolean;
  divided?: boolean;
}) {
  return (
    <header
      className={cn(
        "flex flex-col gap-2 px-4 py-3 lg:flex-row lg:items-center lg:justify-between lg:px-0 lg:py-0 lg:pb-4",
        divided && "border-b",
        compact
          ? "lg:mb-2 lg:border-b-0 lg:pb-2"
          : "lg:mb-4",
      )}
    >
      <div className="min-w-0">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <h1 className="min-w-0 truncate text-xl font-bold lg:text-2xl">
            {title}
          </h1>
          {badge}
        </div>
        {description ? (
          <p className="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">
            {description}
          </p>
        ) : null}
      </div>
      {actions ? (
        <div className="flex w-full shrink-0 flex-wrap items-center gap-2 lg:w-auto">
          {actions}
        </div>
      ) : null}
    </header>
  );
}
