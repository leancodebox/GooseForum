import type { UserBadgePayload } from "@gooseforum/client";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { cn } from "@gooseforum/ui/lib/utils";
import { badgeTone } from "./profile-badge";

export function ProfileAvatar({
  src,
  name,
  badge,
  className,
  framed = true,
  compactBadge = false,
}: {
  src: string;
  name: string;
  badge?: UserBadgePayload | null;
  className?: string;
  framed?: boolean;
  compactBadge?: boolean;
}) {
  return (
    <span className={cn("relative inline-block shrink-0", className)}>
      <Avatar
        className={cn(
          "size-full",
          framed && "border-2 border-background bg-background shadow-sm",
        )}
      >
        <AvatarImage src={src} alt={name} className="object-cover" />
        <AvatarFallback className="text-lg font-semibold">
          {name.trim().slice(0, 1).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      {badge ? (
        <span
          className={cn(
            "absolute -bottom-1 -right-1 z-10 flex size-[38%] min-h-5 min-w-5 items-center justify-center rounded-full p-0.5 shadow-sm ring-1 ring-inset",
            compactBadge && "min-h-4 min-w-4",
            badgeTone(badge.color, badge.level),
          )}
          title={badge.description || badge.name}
        >
          <img
            src={badge.iconUrl || "/static/badges/contributor.svg"}
            alt={badge.name}
            className="size-full object-contain"
          />
        </span>
      ) : null}
    </span>
  );
}
