import { memo, useCallback, useEffect, useRef, useState } from "react";
import type { PagePayload, TopicPayload } from "@gooseforum/client";
import type { TFunction } from "i18next";
import {
  ChevronLeft,
  ChevronRight,
  Grid3X3,
  List,
  Pin,
  Sparkles,
} from "lucide-react";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { Button } from "@gooseforum/ui/components/button";
import { GooseLink, useGoosePageFetcher } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { UserCardPopover } from "../users/user-card-popover";

export type TopicListMode = "waterfall" | "pagination";

type TopicPagination = {
  page: number;
  nextPage: number;
  hasNext: boolean;
  nextUrl: string;
};

type TopicPage = {
  topics: TopicPayload[];
  pagination: TopicPagination;
};

const modeKey = "goose:topic-list-mode";

export function useTopicList<T extends TopicPage>(page: T, pageUrl: string) {
  const fetchPage = useGoosePageFetcher();
  const serverError = useServerErrorMessage();
  const [topics, setTopics] = useState(page.topics);
  const [pagination, setPagination] = useState(page.pagination);
  const [mode, setMode] = useState<TopicListMode>(readMode);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const sentinel = useRef<HTMLDivElement>(null);
  const revision = useRef(0);
  const active = useRef(true);
  const currentPageUrl = useRef(pageUrl);
  const currentTopics = useRef(page.topics);
  const currentPagination = useRef(page.pagination);

  useEffect(() => {
    active.current = true;
    setLoading(false);
    return () => {
      active.current = false;
      revision.current++;
    };
  }, []);

  useEffect(() => {
    if (
      currentPageUrl.current === pageUrl &&
      currentTopics.current === page.topics &&
      currentPagination.current === page.pagination
    )
      return;
    currentPageUrl.current = pageUrl;
    currentTopics.current = page.topics;
    currentPagination.current = page.pagination;
    revision.current++;
    setTopics(page.topics);
    setPagination(page.pagination);
    setLoading(false);
    setError("");
  }, [page, pageUrl]);

  useEffect(() => {
    const unseenById = new Map(
      page.topics.map((topic) => [topic.id, topic.unseen]),
    );
    setTopics((current) => {
      const next = current.map((topic) =>
        unseenById.has(topic.id) && topic.unseen !== unseenById.get(topic.id)
          ? { ...topic, unseen: unseenById.get(topic.id) }
          : topic,
      );
      return next.some((topic, index) => topic !== current[index]) ? next : current;
    });
  }, [page.topics]);

  const loadMore = useCallback(async () => {
    if (
      !active.current ||
      mode !== "waterfall" ||
      loading ||
      !pagination.hasNext ||
      !pagination.nextUrl ||
      !fetchPage
    ) {
      return;
    }
    const current = revision.current;
    setLoading(true);
    setError("");
    try {
      const next = (await fetchPage(
        pagination.nextUrl,
      )) as PagePayload<T>;
      if (!active.current || current !== revision.current) return;
      setTopics((existing) => appendUnique(existing, next.props.topics));
      setPagination(next.props.pagination);
    } catch (reason) {
      if (active.current && current === revision.current) {
      setError(serverError(reason, "load failed"));
      }
    } finally {
      if (active.current && current === revision.current) setLoading(false);
    }
  }, [loading, mode, pagination, fetchPage, serverError]);

  useEffect(() => {
    if (
      mode !== "waterfall" ||
      !sentinel.current ||
      !("IntersectionObserver" in window)
    ) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) void loadMore();
      },
      { rootMargin: "480px 0px" },
    );
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [loadMore, mode]);

  function switchMode() {
    const next = mode === "waterfall" ? "pagination" : "waterfall";
    setMode(next);
    try {
      localStorage.setItem(modeKey, next);
    } catch {}
    if (next === "pagination") {
      setTopics(page.topics);
      setPagination(page.pagination);
    }
  }

  return {
    topics,
    pagination,
    mode,
    loading,
    error,
    sentinel,
    loadMore,
    switchMode,
  };
}

export function TopicListModeSwitch({
  mode,
  t,
  onClick,
}: {
  mode: TopicListMode;
  t: TFunction;
  onClick(): void;
}) {
  const target = mode === "waterfall" ? "pagination" : "waterfall";
  return (
    <span className="border-l pl-2">
      <Button
        variant="ghost"
        size="icon-sm"
        title={t("switchMode", { mode: t(target) })}
        aria-label={t("switchMode", { mode: t(target) })}
        onClick={onClick}
      >
        {mode === "waterfall" ? <List /> : <Grid3X3 />}
      </Button>
    </span>
  );
}

export const TopicTable = memo(function TopicTable({
  topics,
  showPinned = false,
  showCategories = true,
  showHot = true,
  t,
}: {
  topics: TopicPayload[];
  showPinned?: boolean;
  showCategories?: boolean;
  showHot?: boolean;
  t: TFunction;
}) {
  return (
    <div role="table" aria-label={t("topic")}>
      <div
        role="row"
        className="hidden grid-cols-[minmax(0,1fr)_112px_72px_72px_96px] gap-3 border-b px-4 py-1.5 text-[13px] text-muted-foreground lg:grid"
      >
        <span role="columnheader">{t("topic")}</span>
        <span role="columnheader" className="whitespace-nowrap text-center">{t("users")}</span>
        <span role="columnheader" className="whitespace-nowrap text-center">{t("replies")}</span>
        <span role="columnheader" className="whitespace-nowrap text-center">{t("views")}</span>
        <span role="columnheader" className="whitespace-nowrap text-right">{t("activity")}</span>
      </div>
      {topics.map((topic) => (
        <TopicRow
          key={topic.id}
          topic={topic}
          showPinned={showPinned}
          showCategories={showCategories}
          showHot={showHot}
          t={t}
        />
      ))}
    </div>
  );
});

const TopicRow = memo(function TopicRow({
  topic,
  showPinned,
  showCategories,
  showHot,
  t,
}: {
  topic: TopicPayload;
  showPinned: boolean;
  showCategories: boolean;
  showHot: boolean;
  t: TFunction;
}) {
  const isHot = showHot && topic.viewCount > 500;
  const hasMobileLabels =
    (showCategories && topic.categories.length > 0) || isHot;

  return (
    <div
      role="row"
      data-slot="topic-row"
      className="group relative grid grid-cols-[40px_minmax(0,1fr)_auto] grid-rows-[auto_auto] gap-x-2.5 gap-y-1 px-3 py-2 after:absolute after:inset-x-3 after:bottom-0 after:h-px after:bg-border hover:bg-muted/50 last:after:hidden sm:grid-cols-[44px_minmax(0,1fr)_auto] sm:px-4 sm:after:inset-x-4 lg:min-h-16 lg:grid-cols-[minmax(0,1fr)_112px_72px_72px_96px] lg:grid-rows-1 lg:items-center lg:gap-3 lg:py-2.5"
    >
      <div
        role="cell"
        data-slot="topic-row-mobile-avatar"
        className="row-span-2 self-start lg:hidden"
      >
        <UserCardPopover user={topic.author}>
          <GooseLink
            href={`/u/${topic.author.id}`}
            title={topic.author.username}
            className="block size-10 rounded-full sm:size-11"
          >
            <Avatar className="size-full after:hidden">
              <AvatarImage
                src={topic.author.avatarUrl}
                alt={topic.author.username}
              />
              <AvatarFallback>
                {topic.author.username.slice(0, 1)}
              </AvatarFallback>
            </Avatar>
          </GooseLink>
        </UserCardPopover>
      </div>
      <div role="cell" className="min-w-0 lg:col-start-1">
        <div className="flex min-h-6 flex-wrap items-center gap-x-2 gap-y-1">
          <h2 className="flex w-full min-w-0 max-w-full items-start gap-2 lg:w-auto lg:items-center">
            <GooseLink
              href={topic.url}
              data-slot="topic-row-title"
              className="line-clamp-2 min-w-0 text-base font-medium leading-[1.35] group-hover:text-primary sm:text-[17px] lg:block lg:truncate lg:text-base lg:leading-6"
            >
              {showPinned && topic.pinWeight > 0 ? (
                <Pin
                  data-slot="topic-row-pin"
                  aria-hidden="true"
                  className="mr-2 inline size-[1em] rotate-45 align-[-0.125em] text-destructive"
                />
              ) : null}
              {topic.title}
            </GooseLink>
            {topic.unseen ? (
              <span className="mt-1.5 size-2 shrink-0 rounded-full bg-primary lg:mt-0" />
            ) : null}
          </h2>
          {showCategories
            ? topic.categories.map((category) => (
                <GooseLink
                  key={category.id}
                  href={category.url}
                  className="hidden shrink-0 items-center gap-1.5 rounded-full border bg-muted px-2.5 py-1 text-[11px] font-semibold leading-none lg:inline-flex"
                >
                  <span
                    className="size-2 rounded-full"
                    style={{ backgroundColor: category.color }}
                  />
                  {category.name}
                </GooseLink>
              ))
            : null}
          {isHot ? (
            <span className="hidden items-center gap-1 text-[11px] font-semibold text-warning lg:flex">
              <Sparkles className="size-3" />
              hot
            </span>
          ) : null}
        </div>
        <p className="mt-1 hidden min-h-5 truncate text-[13px] text-muted-foreground lg:block">
          {topicDescription(topic)}
        </p>
      </div>
      <div
        role="cell"
        data-slot="topic-row-mobile-replies"
        className="self-start text-right text-base font-semibold tabular-nums lg:hidden"
        aria-label={`${t("replies")}: ${topic.replyCount}`}
      >
        {compactNumber(topic.replyCount)}
      </div>
      <div
        role="cell"
        data-slot="topic-row-mobile-meta"
        className="col-start-2 flex min-w-0 items-center gap-1 overflow-hidden text-xs text-muted-foreground lg:hidden"
      >
        {hasMobileLabels ? (
          <span className="flex min-w-0 items-center gap-1 overflow-hidden">
            {showCategories
              ? topic.categories.map((category) => (
                  <GooseLink
                    key={category.id}
                    href={category.url}
                    className="inline-flex h-5 max-w-28 shrink-0 items-center gap-1 truncate rounded-md bg-muted px-1.5 text-[11px] font-semibold leading-none"
                  >
                    <span
                      className="size-1.5 shrink-0 rounded-full"
                      style={{ backgroundColor: category.color }}
                    />
                    <span className="truncate">{category.name}</span>
                  </GooseLink>
                ))
              : null}
            {isHot ? (
              <span className="inline-flex shrink-0 items-center gap-1 text-[11px] font-semibold text-warning">
                <Sparkles className="size-3" />
                hot
              </span>
            ) : null}
          </span>
        ) : null}
      </div>
      <time
        role="cell"
        className="col-start-3 whitespace-nowrap self-end text-right text-[11px] text-muted-foreground lg:hidden"
        dateTime={topic.lastUpdateTime}
      >
        {relativeTime(topic.lastUpdateTime, t)}
      </time>
      <div role="cell" className="hidden text-xs text-muted-foreground lg:flex lg:justify-center">
        <AvatarStack users={topic.participants} />
      </div>
      <div role="cell" className="hidden text-center font-semibold lg:block">
        {compactNumber(topic.replyCount)}
      </div>
      <div
        role="cell"
        className="hidden text-center text-muted-foreground lg:block"
      >
        {compactNumber(topic.viewCount)}
      </div>
      <div
        role="cell"
        className="hidden text-right text-[13px] text-muted-foreground lg:block"
      >
        {relativeTime(topic.lastUpdateTime, t)}
      </div>
    </div>
  );
});

function AvatarStack({
  users,
}: {
  users: TopicPayload["participants"];
}) {
  return (
    <div
      data-slot="avatar-stack"
      className="relative h-[var(--avatar-size)] shrink-0 [--avatar-size:24px] [--avatar-step:16px] lg:[--avatar-size:32px] lg:[--avatar-step:20px]"
      style={{ width: `calc(var(--avatar-size) + ${Math.max(0, users.length - 1)} * var(--avatar-step))` }}
    >
      {users.map((user, index) => (
        <UserCardPopover key={user.id} user={user}>
          <GooseLink
            href={`/u/${user.id}`}
            title={user.username}
            style={{ left: `calc(${index} * var(--avatar-step))` }}
            className="absolute inset-y-0 size-[var(--avatar-size)] rounded-full ring-2 ring-background transition-transform duration-150 hover:z-10 hover:scale-110 data-[state=open]:z-10 data-[state=open]:scale-110"
          >
            <Avatar className="size-full after:hidden">
              <AvatarImage src={user.avatarUrl} alt={user.username} />
              <AvatarFallback>{user.username.slice(0, 1)}</AvatarFallback>
            </Avatar>
          </GooseLink>
        </UserCardPopover>
      ))}
    </div>
  );
}

export function TopicListFooter({
  mode,
  pagination,
  loading,
  error,
  t,
  onLoad,
  previousUrl,
}: {
  mode: TopicListMode;
  pagination: TopicPagination;
  loading: boolean;
  error: string;
  t: TFunction;
  onLoad(): Promise<void>;
  previousUrl?: string;
}) {
  const previous =
    previousUrl ?? (pagination.page > 1 ? `?page=${pagination.page - 1}` : "");
  return (
    <nav className="border-t bg-muted/50 p-3 text-center">
      {mode === "pagination" ? (
        <div className="flex justify-center gap-2">
          {previous ? (
            <Button asChild variant="outline" size="sm">
              <GooseLink href={previous} rel="prev">
                <ChevronLeft />
                {t("previous")}
              </GooseLink>
            </Button>
          ) : null}
          <span className="px-2 py-1.5 text-xs font-semibold">
            {t("currentPage", { page: pagination.page })}
          </span>
          {pagination.hasNext ? (
            <Button asChild variant="outline" size="sm">
              <GooseLink href={pagination.nextUrl} rel="next">
                {t("next")}
                <ChevronRight />
              </GooseLink>
            </Button>
          ) : null}
        </div>
      ) : pagination.hasNext ? (
        <Button
          variant="ghost"
          size="sm"
          disabled={loading}
          onClick={() => void onLoad()}
        >
          {loading ? t("loading") : t("loadMore")}
        </Button>
      ) : (
        <p className="text-xs text-muted-foreground">{t("allShown")}</p>
      )}
      {error ? (
        <p className="mt-2 text-xs text-destructive">{t("autoLoadFailed")}</p>
      ) : null}
    </nav>
  );
}

export function compactNumber(value: number) {
  return value >= 1e6
    ? `${(value / 1e6).toFixed(value >= 1e7 ? 0 : 1)}m`
    : value >= 1e3
      ? `${(value / 1e3).toFixed(value >= 1e4 ? 0 : 1)}k`
      : String(value);
}

function readMode(): TopicListMode {
  try {
    return localStorage.getItem(modeKey) === "pagination"
      ? "pagination"
      : "waterfall";
  } catch {
    return "waterfall";
  }
}

function appendUnique(existing: TopicPayload[], incoming: TopicPayload[]) {
  const ids = new Set(existing.map((topic) => topic.id));
  return [
    ...existing,
    ...incoming.filter((topic) => {
      if (ids.has(topic.id)) return false;
      ids.add(topic.id);
      return true;
    }),
  ];
}

function topicDescription(topic: TopicPayload) {
  return (
    topic.description?.trim() ||
    [
      "(｀・ω・´)",
      "( ´ ▽ ` )ﾉ",
      "(ง •̀_•́)ง",
      "(｡･ω･｡)",
      "(￣▽￣)ノ",
      "(っ´ω`)っ",
    ][Math.abs(topic.id) % 6]
  );
}

function relativeTime(value: string, t: TFunction) {
  const time = new Date(
    value.includes("T") ? value : value.replace(" ", "T"),
  ).getTime();
  if (!Number.isFinite(time)) return value;
  const seconds = Math.max(0, Math.floor((Date.now() - time) / 1000));
  if (seconds < 60) return t("justNow");
  if (seconds < 3600)
    return t("minuteAgo", { count: Math.floor(seconds / 60) });
  if (seconds < 86400)
    return t("hourAgo", { count: Math.floor(seconds / 3600) });
  if (seconds < 604800)
    return t("dayAgo", { count: Math.floor(seconds / 86400) });
  return value.slice(0, 10);
}
