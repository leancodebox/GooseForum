import type { UserBadgePayload } from "@gooseforum/client";
import { cn } from "@gooseforum/ui/lib/utils";

export function ProfileBadge({
  badge,
  compact = false,
}: {
  badge: UserBadgePayload;
  compact?: boolean;
}) {
  return (
    <span
      className={cn(
        "flex shrink-0 items-center justify-center ring-1 ring-inset",
        compact ? "size-10" : "size-11",
        badgeTone(badge.color, badge.level),
      )}
      style={{
        clipPath: "polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)",
      }}
      title={badge.description || badge.name}
    >
      <img
        src={badge.iconUrl || "/static/badges/contributor.svg"}
        alt={badge.name}
        className="size-5 object-contain"
      />
    </span>
  );
}

export function badgeTone(color: string, level: string) {
  const tones: Record<string, string> = {
    blue: "bg-blue-100 text-blue-700 ring-blue-200",
    emerald: "bg-emerald-100 text-emerald-700 ring-emerald-200",
    teal: "bg-teal-100 text-teal-700 ring-teal-200",
    sky: "bg-sky-100 text-sky-700 ring-sky-200",
    cyan: "bg-cyan-100 text-cyan-700 ring-cyan-200",
    rose: "bg-rose-100 text-rose-700 ring-rose-200",
    violet: "bg-violet-100 text-violet-700 ring-violet-200",
    purple: "bg-purple-100 text-purple-700 ring-purple-200",
    fuchsia: "bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200",
    indigo: "bg-indigo-100 text-indigo-700 ring-indigo-200",
    amber: "bg-amber-100 text-amber-700 ring-amber-200",
    orange: "bg-orange-100 text-orange-700 ring-orange-200",
    yellow: "bg-yellow-100 text-yellow-700 ring-yellow-200",
    slate: "bg-slate-100 text-slate-700 ring-slate-200",
  };
  return (
    tones[color] ||
    (level === "gold"
      ? tones.amber
      : level === "special"
        ? tones.indigo
        : tones.blue)
  );
}
