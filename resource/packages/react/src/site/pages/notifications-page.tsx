import { useCallback, useEffect, useRef, useState } from "react";
import type {
  NotificationFilter,
  NotificationPayload,
  NotificationsPageProps,
} from "@gooseforum/client";
import {
  Award,
  Bell,
  Check,
  CheckCheck,
  Info,
  MessageCircle,
  UserPlus,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { Tabs, TabsList, TabsTrigger } from "../../components/ui/tabs";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { announceUnreadStatus } from "../../runtime/unread-status";
import { PageHeader } from "../layout/page-header";
import { UserCardPopover } from "../users/user-card-popover";

type ListState = {
  items: NotificationPayload[];
  nextCursor: number;
  hasNext: boolean;
  loading: boolean;
  loaded: boolean;
};

export function NotificationsPageView({
  page,
}: {
  page: NotificationsPageProps;
}) {
  const { t } = useTranslation("notifications");
  const runtime = useGooseRuntime();
  const [filter, setFilter] = useState<NotificationFilter>("all");
  const [unreadCount, setUnreadCount] = useState(page.unreadCount);
  const [markingAll, setMarkingAll] = useState(false);
  const [error, setError] = useState("");
  const [lists, setLists] = useState<Record<NotificationFilter, ListState>>(
    () => ({
      all: {
        items: page.notifications.map(copy),
        nextCursor: initialCursor(page),
        hasNext: page.pagination.hasNext,
        loading: false,
        loaded: true,
      },
      unread: {
        items: [],
        nextCursor: 0,
        hasNext: true,
        loading: false,
        loaded: false,
      },
    }),
  );
  const sentinel = useRef<HTMLDivElement>(null);
  const active = lists[filter];

  const load = useCallback(
    async (target: NotificationFilter, reset = false) => {
      const current = lists[target];
      if (current.loading || (!reset && !current.hasNext)) return;
      setError("");
      setLists((state) => ({
        ...state,
        [target]: { ...state[target], loading: true },
      }));
      try {
        const response = await runtime.api.notifications.list(
          target,
          reset ? 0 : current.nextCursor,
          20,
        );
        setUnreadCount(response.unreadCount);
        announceUnreadStatus({ notifications: response.unreadCount > 0 });
        setLists((state) => ({
          ...state,
          [target]: {
            items: reset
              ? response.items.map(copy)
              : merge(state[target].items, response.items),
            nextCursor: response.nextCursor,
            hasNext: response.hasNext,
            loading: false,
            loaded: true,
          },
        }));
      } catch (reason) {
        setError(reason instanceof Error ? reason.message : t("loadFailed"));
        setLists((state) => ({
          ...state,
          [target]: { ...state[target], loading: false },
        }));
      }
    },
    [lists, runtime.api.notifications, t],
  );

  useEffect(() => {
    if (filter === "unread" && !lists.unread.loaded) void load("unread", true);
  }, [filter, lists.unread.loaded, load]);

  useEffect(() => {
    if (
      !sentinel.current ||
      !active.hasNext ||
      !("IntersectionObserver" in window)
    )
      return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) void load(filter);
      },
      { rootMargin: "160px 0px" },
    );
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [active.hasNext, filter, load]);

  async function markAllRead() {
    if (!unreadCount || markingAll) return;
    const previous = lists;
    const previousCount = unreadCount;
    setMarkingAll(true);
    setError("");
    setUnreadCount(0);
    announceUnreadStatus({ notifications: false });
    setLists((state) => ({
      all: {
        ...state.all,
        items: state.all.items.map((item) => ({ ...item, isRead: true })),
      },
      unread: { ...state.unread, items: [], hasNext: false, loaded: true },
    }));
    try {
      await runtime.api.notifications.markAllRead();
    } catch (reason) {
      setLists(previous);
      setUnreadCount(previousCount);
      announceUnreadStatus({ notifications: previousCount > 0 });
      setError(
        reason instanceof Error ? reason.message : t("markAllReadFailed"),
      );
    } finally {
      setMarkingAll(false);
    }
  }

  function markRead(item: NotificationPayload) {
    if (item.isRead) return;
    const previous = lists;
    const previousCount = unreadCount;
    setUnreadCount((count) => Math.max(0, count - 1));
    announceUnreadStatus({ notifications: previousCount - 1 > 0 });
    setLists((state) => ({
      all: {
        ...state.all,
        items: state.all.items.map((entry) =>
          entry.id === item.id ? { ...entry, isRead: true } : entry,
        ),
      },
      unread: {
        ...state.unread,
        items: state.unread.items.filter((entry) => entry.id !== item.id),
      },
    }));
    void runtime.api.notifications.markRead(item.id).catch(() => {
      setLists(previous);
      setUnreadCount(previousCount);
      announceUnreadStatus({ notifications: previousCount > 0 });
    });
  }

  return (
    <main className="min-w-0 pb-8">
      <PageHeader
        title={t("title")}
        description={t("summary", { total: active.items.length })}
        badge={
          unreadCount ? (
            <Badge>{t("unread", { count: unreadCount })}</Badge>
          ) : null
        }
        actions={
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!unreadCount || markingAll}
            onClick={() => void markAllRead()}
          >
            <CheckCheck data-icon="inline-start" />
            {markingAll ? t("loadingMore") : t("markAllRead")}
          </Button>
        }
      />
      {error ? (
        <p className="px-4 py-2 text-sm text-destructive lg:px-0">{error}</p>
      ) : null}
      <section className="overflow-hidden rounded-xl border bg-background">
        <Tabs
          value={filter}
          onValueChange={(value) => setFilter(value as NotificationFilter)}
          className="gap-0"
        >
          <TabsList className="h-auto w-full justify-start rounded-none border-b bg-muted/50 p-2">
            <TabsTrigger value="all" className="flex-none px-3">
              {t("tabs.all")}
            </TabsTrigger>
            <TabsTrigger value="unread" className="flex-none px-3">
              {t("tabs.unread")}
              {unreadCount ? (
                <Badge className="px-1.5">{unreadCount}</Badge>
              ) : null}
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="hidden grid-cols-[34px_minmax(0,1fr)_116px] gap-3 border-b bg-muted/50 px-3 py-2 text-[11px] font-bold uppercase text-muted-foreground md:grid">
          <span />
          <span>{t("notification")}</span>
          <span className="text-right">{t("time")}</span>
        </div>
        {active.items.length ? (
          <div className="divide-y">
            {active.items.map((item) => (
              <NotificationRow
                key={item.id}
                item={item}
                locale={runtime.locale}
                t={t}
                onRead={() => markRead(item)}
              />
            ))}
          </div>
        ) : active.loading ? (
          <NotificationEmpty loading title={t("loadingMore")} description="" />
        ) : (
          <NotificationEmpty
            title={t(filter === "unread" ? "unreadEmptyTitle" : "emptyTitle")}
            description={t(
              filter === "unread"
                ? "unreadEmptyDescription"
                : "emptyDescription",
            )}
          />
        )}
        <div
          ref={sentinel}
          className="border-t px-4 py-3 text-center text-xs font-semibold text-muted-foreground"
        >
          {active.hasNext ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={active.loading}
              onClick={() => void load(filter)}
            >
              {active.loading ? t("loadingMore") : t("loadMore")}
            </Button>
          ) : active.items.length ? (
            t("noMore")
          ) : null}
        </div>
      </section>
    </main>
  );
}

function NotificationRow({
  item,
  locale,
  t,
  onRead,
}: {
  item: NotificationPayload;
  locale: string;
  t: ReturnType<typeof useTranslation>["t"];
  onRead(): void;
}) {
  const Icon =
    item.eventType === "follow"
      ? UserPlus
      : item.eventType === "badge"
        ? Award
        : item.eventType === "system"
          ? Info
          : MessageCircle;
  const actor = actorName(item, t);
  const actorUrl = item.actor.id ? `/u/${item.actor.id}` : "";
  const target = targetUrl(item, actorUrl);
  return (
    <article
      className={
        item.isRead
          ? "relative grid grid-cols-[34px_minmax(0,1fr)] gap-3 px-3 py-2.5 hover:bg-muted/40 md:grid-cols-[34px_minmax(0,1fr)_116px_40px]"
          : "relative grid grid-cols-[34px_minmax(0,1fr)] gap-3 bg-primary/5 px-3 py-2.5 before:absolute before:inset-y-0 before:left-0 before:w-0.5 before:bg-primary hover:bg-primary/10 md:grid-cols-[34px_minmax(0,1fr)_116px_40px]"
      }
    >
      <span className="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
        <Icon className="size-4" />
      </span>
      <div className="min-w-0">
        <div className="flex min-w-0 items-center gap-1.5 text-sm leading-5">
          {actorUrl && item.eventType !== "badge" ? (
            <UserCardPopover
              user={{
                id: item.actor.id,
                username: actor,
                avatarUrl: item.actor.avatarUrl,
              }}
            >
              <GooseLink
                href={actorUrl}
                className="max-w-[42%] shrink-0 truncate font-semibold hover:text-primary"
                onClick={onRead}
              >
                {actor}
              </GooseLink>
            </UserCardPopover>
          ) : (
            <span className="max-w-[42%] shrink-0 truncate font-semibold">
              {item.eventType === "follow" ? actor : titleText(item, t)}
            </span>
          )}
          {item.actor.id || item.eventType === "follow" ? (
            <span className="shrink-0 text-muted-foreground">
              {verb(item, t)}
            </span>
          ) : null}
          {target ? (
            <GooseLink
              href={target}
              className="min-w-0 truncate font-semibold text-primary"
              onClick={onRead}
            >
              {notificationText(item, t)}
            </GooseLink>
          ) : (
            <span className="min-w-0 truncate font-medium">
              {notificationText(item, t)}
            </span>
          )}
          {!item.isRead ? (
            <span className="size-1.5 rounded-full bg-primary" />
          ) : null}
        </div>
        {item.content && item.content !== notificationText(item, t) ? (
          <p className="line-clamp-1 text-xs text-muted-foreground">
            {item.content}
          </p>
        ) : null}
        <time className="mt-1 block text-xs text-muted-foreground md:hidden">
          {formatDate(item.createdAt, locale)}
        </time>
      </div>
      <time className="hidden text-right text-xs text-muted-foreground md:block">
        {formatDate(item.createdAt, locale)}
      </time>
      {!item.isRead ? (
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="absolute right-2 top-2 md:static"
          title={t("markRead")}
          aria-label={t("markRead")}
          onClick={onRead}
        >
          <Check />
        </Button>
      ) : null}
    </article>
  );
}

function NotificationEmpty({
  title,
  description,
  loading = false,
}: {
  title: string;
  description: string;
  loading?: boolean;
}) {
  return (
    <Empty className="min-h-52 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Bell className={loading ? "animate-pulse" : ""} />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        {description ? (
          <EmptyDescription>{description}</EmptyDescription>
        ) : null}
      </EmptyHeader>
    </Empty>
  );
}

function copy(item: NotificationPayload) {
  return { ...item };
}
function merge(
  current: NotificationPayload[],
  incoming: NotificationPayload[],
) {
  const ids = new Set(current.map((item) => item.id));
  return [
    ...current,
    ...incoming.filter((item) => {
      if (ids.has(item.id)) return false;
      ids.add(item.id);
      return true;
    }),
  ];
}
function initialCursor(page: NotificationsPageProps) {
  if (!page.pagination.hasNext) return 0;
  try {
    return (
      Number(
        new URL(
          page.pagination.nextUrl,
          "http://gooseforum.local",
        ).searchParams.get("cursor"),
      ) ||
      page.notifications.at(-1)?.id ||
      0
    );
  } catch {
    return page.notifications.at(-1)?.id || 0;
  }
}
function template(
  item: NotificationPayload,
  t: ReturnType<typeof useTranslation>["t"],
) {
  const key = item.payload.templateKey?.replace("notifications.templates.", "");
  return key
    ? t(`templates.${key}`, {
        badge:
          item.payload.templateParams?.badgeName ||
          item.payload.metadata?.badgeName ||
          "",
      })
    : "";
}
function actorName(
  item: NotificationPayload,
  t: ReturnType<typeof useTranslation>["t"],
) {
  return item.eventType === "badge"
    ? template(item, t) || item.title || t("actorFallback")
    : item.actor.username ||
        item.payload.actorName ||
        item.payload.metadata?.followerName ||
        t("actorFallback");
}
function titleText(
  item: NotificationPayload,
  t: ReturnType<typeof useTranslation>["t"],
) {
  return item.title || t("fallback");
}
function notificationText(
  item: NotificationPayload,
  t: ReturnType<typeof useTranslation>["t"],
) {
  const translated = template(item, t);
  if (item.eventType === "badge")
    return (
      translated ||
      (item.payload.metadata?.badgeName
        ? t("badgeEarned", { badge: item.payload.metadata.badgeName })
        : item.content || item.payload.content || t("fallback"))
    );
  if (item.eventType === "follow")
    return (
      translated ||
      item.content ||
      item.payload.content ||
      t("followDescription", { actor: actorName(item, t) })
    );
  return (
    item.topic?.title ||
    translated ||
    item.content ||
    item.payload.content ||
    t("fallback")
  );
}
function verb(
  item: NotificationPayload,
  t: ReturnType<typeof useTranslation>["t"],
) {
  const translated = template(item, t);
  if (translated && item.eventType !== "badge") return translated;
  if (item.eventType === "follow") return t("verb.follow");
  if (item.eventType === "post_reply") return t("verb.reply");
  if (item.eventType === "comment" || item.eventType === "topic_post")
    return t("verb.comment");
  return titleText(item, t);
}
function targetUrl(item: NotificationPayload, actorUrl: string) {
  if (item.topic) return item.topic.url;
  if (item.eventType === "badge")
    return item.payload.metadata?.profileUrl || actorUrl;
  if (item.eventType === "follow") return actorUrl;
  return "";
}
function formatDate(value: string, locale: string) {
  const date = new Date(value.includes("T") ? value : value.replace(" ", "T"));
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });
}
