import { useRef, useState, type MouseEvent, type ReactElement } from "react";
import type {
  GooseSiteApi,
  UserBadgePayload,
  UserCardPayload,
} from "@gooseforum/client";
import { formatCompactNumber } from "@gooseforum/client";
import { CalendarDays, Radio, UserPlus } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../../components/ui/popover";
import { Skeleton } from "../../components/ui/skeleton";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { ProfileAvatar } from "./profile-avatar";
import { ProfileBadge } from "./profile-badge";

export interface UserCardTarget {
  id: number;
  username: string;
  avatarUrl?: string;
  wornBadge?: UserBadgePayload | null;
}

const cardCache = new Map<number, UserCardPayload>();
const pendingCards = new Map<number, Promise<UserCardPayload>>();

export function UserCardPopover({
  user,
  children,
}: {
  user: UserCardTarget;
  children: ReactElement;
}) {
  const { t } = useTranslation("userCard");
  const runtime = useGooseRuntime();
  const [open, setOpen] = useState(false);
  const [opened, setOpened] = useState(false);
  const [card, setCard] = useState(() => cardCache.get(user.id));
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const ignoreNextOpen = useRef(false);

  function captureClick(event: MouseEvent) {
    if (
      event.button !== 0 ||
      event.metaKey ||
      event.ctrlKey ||
      event.shiftKey ||
      event.altKey
    ) {
      ignoreNextOpen.current = true;
      return;
    }
    event.preventDefault();
    openChange(!open);
  }

  function openChange(next: boolean) {
    if (next && ignoreNextOpen.current) {
      ignoreNextOpen.current = false;
      return;
    }
    ignoreNextOpen.current = false;
    setOpen(next);
    if (!next) return;
    setOpened(true);
    if (card || loading) return;
    setLoading(true);
    setFailed(false);
    void loadCard(runtime.api, user.id)
      .then(setCard)
      .catch(() => setFailed(true))
      .finally(() => setLoading(false));
  }

  const profileUrl = `/u/${card?.userId || user.id}`;
  const username = card?.username || user.username;
  const displayName = card?.nickname || username;
  return (
    <Popover open={open} onOpenChange={openChange}>
      <PopoverTrigger asChild onClickCapture={captureClick}>
        {children}
      </PopoverTrigger>
      {opened ? (
        <PopoverContent
          align="start"
          sideOffset={10}
          className="max-h-[var(--radix-popover-content-available-height)] w-[min(20rem,calc(100vw-1.5rem))] overflow-y-auto p-3"
          aria-label={displayName}
        >
          <div className="flex items-start gap-3">
            <GooseLink href={profileUrl} className="shrink-0 rounded-full">
              <ProfileAvatar
                src={card?.avatarUrl || user.avatarUrl || ""}
                name={username}
                badge={card?.wornBadge || user.wornBadge}
                className="size-14"
              />
            </GooseLink>
            <div className="min-w-0 flex-1">
              <div className="flex min-w-0 items-center gap-2">
                <GooseLink
                  href={profileUrl}
                  className="truncate text-base font-bold hover:text-primary"
                >
                  {displayName}
                </GooseLink>
                {card?.isAdmin ? (
                  <Badge variant="secondary" className="text-warning">
                    Admin
                  </Badge>
                ) : null}
              </div>
              <div className="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
                <span className="truncate">@{username}</span>
                {card?.isOnline ? (
                  <span className="inline-flex items-center gap-1 text-success">
                    <Radio className="size-3" />
                    {t("online")}
                  </span>
                ) : card?.lastActiveTime ? (
                  <span>
                    {t("activeAt", {
                      time: relativeTime(card.lastActiveTime, runtime.locale),
                    })}
                  </span>
                ) : null}
              </div>
            </div>
          </div>
          {loading ? (
            <UserCardSkeleton label={t("loading")} />
          ) : failed ? (
            <p className="mt-3 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {t("unavailable")}
            </p>
          ) : card ? (
            <>
              {card.bio || card.signature ? (
                <p className="mt-3 line-clamp-2 text-sm leading-relaxed text-foreground/75">
                  {card.bio || card.signature}
                </p>
              ) : null}
              {card.badges.length ? (
                <div className="mt-3 flex gap-2">
                  {card.badges.slice(0, 5).map((badge) => (
                    <ProfileBadge key={badge.code} badge={badge} compact />
                  ))}
                </div>
              ) : null}
              <div className="mt-3 grid grid-cols-4 divide-x border-y py-2">
                {[
                  ["topics", card.topicCount],
                  ["replies", card.replyCount],
                  ["likes", card.likeReceivedCount],
                  ["followers", card.followerCount],
                ].map(([key, value]) => (
                  <div key={key} className="px-1 text-center">
                    <div className="text-sm font-bold tabular-nums">
                      {formatCompactNumber(Number(value))}
                    </div>
                    <div className="text-[11px] text-muted-foreground">
                      {t(`stats.${key}`)}
                    </div>
                  </div>
                ))}
              </div>
              <div className="mt-3 flex items-center justify-between gap-3">
                <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
                  <CalendarDays className="size-3.5" />
                  {t("joinedAt", {
                    date: new Date(card.createdAt).toLocaleDateString(
                      runtime.locale,
                    ),
                  })}
                </span>
                <Button asChild size="sm">
                  <GooseLink href={profileUrl}>
                    <UserPlus data-icon="inline-start" />
                    {card.isFollowing ? t("following") : t("viewProfile")}
                  </GooseLink>
                </Button>
              </div>
            </>
          ) : null}
        </PopoverContent>
      ) : null}
    </Popover>
  );
}

function UserCardSkeleton({ label }: { label: string }) {
  return (
    <div className="mt-3 flex min-h-40 flex-col gap-3">
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-3/4" />
      <div className="grid grid-cols-4 gap-2 border-y py-2">
        {[0, 1, 2, 3].map((key) => (
          <Skeleton key={key} className="h-8" />
        ))}
      </div>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  );
}

function loadCard(api: GooseSiteApi, userId: number) {
  const cached = cardCache.get(userId);
  if (cached) return Promise.resolve(cached);
  const pending = pendingCards.get(userId);
  if (pending) return pending;
  const request = api.users
    .card(userId)
    .then((card) => {
      cardCache.set(userId, card);
      return card;
    })
    .finally(() => pendingCards.delete(userId));
  pendingCards.set(userId, request);
  return request;
}

function relativeTime(value: string, locale: string) {
  const time = new Date(
    value.includes("T") ? value : value.replace(" ", "T"),
  ).getTime();
  if (!Number.isFinite(time)) return value;
  const seconds = Math.round((time - Date.now()) / 1000);
  const formatter = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });
  if (Math.abs(seconds) < 60) return formatter.format(seconds, "second");
  if (Math.abs(seconds) < 3600)
    return formatter.format(Math.round(seconds / 60), "minute");
  if (Math.abs(seconds) < 86400)
    return formatter.format(Math.round(seconds / 3600), "hour");
  return formatter.format(Math.round(seconds / 86400), "day");
}
