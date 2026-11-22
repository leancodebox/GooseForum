import { avatarInitial } from "@gooseforum/ui/lib/avatar-initial";
import type { UserBadgePayload } from "@gooseforum/client";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { cn } from "@gooseforum/ui/lib/utils";
import { ForumBadge } from "@gooseforum/ui/components/forum-badge";

export function ProfileAvatar({
  src,
  name,
  badge,
  className,
  framed = true,
  compactBadge = false,
  badgePosition = "right",
  hideOutline,
  avatarClassName,
  badgeClassName,
}: {
  src: string;
  name: string;
  badge?: UserBadgePayload | null;
  className?: string;
  framed?: boolean;
  compactBadge?: boolean;
  badgePosition?: "left" | "right";
  hideOutline?: boolean;
  avatarClassName?: string;
  badgeClassName?: string;
}) {
  const shouldHideOutline = hideOutline ?? !framed;
  return (
    <span className={cn("relative isolate inline-block shrink-0", className)}>
      <Avatar
        className={cn(
          "size-full @container",
          framed && "border-2 border-background bg-background shadow-sm",
          shouldHideOutline && "after:hidden",
          avatarClassName,
        )}
      >
        <AvatarImage src={src} alt={name} className="object-cover" />
        <AvatarFallback className="text-[length:50cqi] leading-none font-semibold">
          {avatarInitial(name)}
        </AvatarFallback>
      </Avatar>
      {badge ? (
        <ForumBadge
          badge={badge}
          variant="worn"
          className={cn(
            "pointer-events-none absolute -bottom-1 z-30 size-[38%] min-h-5 min-w-5",
            badgePosition === "left" ? "-left-1" : "-right-1",
            compactBadge && "min-h-4 min-w-4",
            badgeClassName,
          )}
        />
      ) : null}
    </span>
  );
}
