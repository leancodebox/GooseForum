import { useState } from "react";
import { cn } from "../lib/utils";

export interface ForumBadgeAppearance {
  name: string;
  description?: string;
  iconUrl?: string;
  color: string;
  level: string;
}

const fallbackIcon = "/static/badges/contributor.svg";

export function ForumBadge({
  badge,
  size = "md",
  variant = "hexagon",
  className,
  onImageError,
}: {
  badge: ForumBadgeAppearance;
  size?: "sm" | "md" | "lg";
  variant?: "hexagon" | "worn";
  className?: string;
  onImageError?: () => void;
}) {
  const source = badge.iconUrl || fallbackIcon;
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center ring-1 ring-inset",
        size === "sm" ? "size-10" : size === "lg" ? "size-16" : "size-11",
        variant === "worn" && "rounded-full p-0.5 shadow-sm",
        badgeTone(badge.color, badge.level),
        className,
      )}
      style={variant === "hexagon" ? {
        clipPath: "polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)",
      } : undefined}
      title={badge.description || badge.name}
    >
      <BadgeImage
        key={source}
        source={source}
        alt={badge.name}
        className={cn("object-contain", variant === "worn" ? "size-full" : size === "lg" ? "size-8" : "size-5")}
        onImageError={onImageError}
      />
    </span>
  );
}

function BadgeImage({ source, alt, className, onImageError }: {
  source: string;
  alt: string;
  className: string;
  onImageError?: () => void;
}) {
  const [failed, setFailed] = useState(false);
  return <img src={failed ? fallbackIcon : source} alt={alt} className={className} onError={() => {
    if (!failed) {
      setFailed(true);
      onImageError?.();
    }
  }} />;
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
  return tones[color] || (level === "gold" ? tones.amber : level === "silver" ? tones.slate : level === "special" ? tones.indigo : tones.blue);
}
