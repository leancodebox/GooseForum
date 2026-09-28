import type { ReactNode } from "react";

export function ProfileIdentity({
  displayName,
  username,
  description,
  badges,
}: {
  displayName: string;
  username: string;
  description: string;
  badges?: ReactNode;
}) {
  return (
    <div className="min-w-0 flex-1 pt-3">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <h1 className="truncate text-2xl font-bold leading-tight">
          {displayName}
        </h1>
        {badges}
      </div>
      <div className="mt-1 flex min-w-0 flex-wrap items-center gap-1.5">
        <p className="truncate text-sm font-medium text-muted-foreground">
          @{username}
        </p>
      </div>
      <p className="mt-2 text-sm leading-relaxed text-foreground/75">
        {description}
      </p>
    </div>
  );
}
