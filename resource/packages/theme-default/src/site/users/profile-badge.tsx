import type { UserBadgePayload } from "@gooseforum/client";
import { ForumBadge } from "@gooseforum/ui/components/forum-badge";

export function ProfileBadge({
  badge,
  compact = false,
}: {
  badge: UserBadgePayload;
  compact?: boolean;
}) {
  return <ForumBadge badge={badge} size={compact ? "sm" : "md"} />;
}
