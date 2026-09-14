import { useCallback, useEffect, useRef, useState } from "react";
import type { PagePayload, TopicPayload } from "@gooseforum/client";
import type { TFunction } from "i18next";
import {
  ChevronLeft,
  ChevronRight,
  Grid3X3,
  List,
  MessageSquare,
  Pin,
  Sparkles,
} from "lucide-react";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import { GooseLink, useGooseRuntime } from "../../runtime";
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
  const runtime = useGooseRuntime();
  const [topics, setTopics] = useState(page.topics);
  const [pagination, setPagination] = useState(page.pagination);
  const [mode, setMode] = useState<TopicListMode>(readMode);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const sentinel = useRef<HTMLDivElement>(null);
  const revision = useRef(0);

  useEffect(() => {
    revision.current++;
    setTopics(page.topics);
    setPagination(page.pagination);
    setLoading(false);
    setError("");
  }, [page, pageUrl]);

  const loadMore = useCallback(async () => {
    if (
      mode !== "waterfall" ||
      loading ||
      !pagination.hasNext ||
      !pagination.nextUrl ||
      !runtime.fetchPage
    ) {
      return;
    }
    const current = revision.current;
    setLoading(true);
    setError("");
    try {
      const next = (await runtime.fetchPage(
        pagination.nextUrl,
      )) as PagePayload<T>;
      if (current !== revision.current) return;
      setTopics((existing) => appendUnique(existing, next.props.topics));
      setPagination(next.props.pagination);
    } catch (reason) {
      if (current === revision.current) {
        setError(reason instanceof Error ? reason.message : "load failed");
      }
    } finally {
      if (current === revision.current) setLoading(false);
    }
  }, [loading, mode, pagination, runtime]);

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

export function TopicTable({
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
        className="hidden grid-cols-[minmax(0,1fr)_112px_72px_72px_88px] gap-3 border-b px-4 py-1.5 text-[13px] text-muted-foreground lg:grid"
      >
        <span>{t("topic")}</span>
        <span className="text-center">{t("users")}</span>
        <span className="text-center">{t("replies")}</span>
        <span className="text-center">{t("views")}</span>
        <span className="text-right">{t("activity")}</span>
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
}

function TopicRow({
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
  return (
    <div
      role="row"
      className="group relative grid min-h-[88px] gap-2.5 px-4 py-2.5 after:absolute after:inset-x-4 after:bottom-0 after:h-px after:bg-border hover:bg-muted/50 last:after:hidden lg:min-h-16 lg:grid-cols-[minmax(0,1fr)_112px_72px_72px_88px] lg:items-center"
    >
      <div role="cell" className="min-w-0">
        <div className="flex min-h-6 flex-wrap items-center gap-x-2 gap-y-1">
          <h2 className="flex min-w-0 max-w-full items-center gap-2">
            {showPinned && topic.pinWeight > 0 ? (
              <Pin className="size-3.5 rotate-45 text-destructive" />
            ) : null}
            <GooseLink
              href={topic.url}
              className="truncate text-[15px] font-medium leading-6 group-hover:text-primary sm:text-base"
            >
              {topic.title}
            </GooseLink>
            {topic.unseen ? (
              <span className="size-2 rounded-full bg-primary" />
            ) : null}
          </h2>
          {showCategories
            ? topic.categories.map((category) => (
                <GooseLink
                  key={category.id}
                  href={category.url}
                  className="inline-flex items-center gap-1.5 rounded-full border bg-muted px-2.5 py-1 text-[11px] font-semibold"
                >
                  <span
                    className="size-2 rounded-full"
                    style={{ backgroundColor: category.color }}
                  />
                  {category.name}
                </GooseLink>
              ))
            : null}
          {showHot && topic.viewCount > 500 ? (
            <span className="flex items-center gap-1 text-[11px] font-semibold text-warning">
              <Sparkles className="size-3" />
              hot
            </span>
          ) : null}
        </div>
        <p className="mt-1 min-h-5 truncate text-[13px] text-muted-foreground">
          {topicDescription(topic)}
        </p>
        <div className="mt-1.5 flex items-center gap-3 text-xs text-muted-foreground lg:hidden">
          <AvatarStack users={topic.participants} />
          <time>{relativeTime(topic.lastUpdateTime, t)}</time>
          <span className="flex items-center gap-1">
            <MessageSquare className="size-3.5" />
            {compactNumber(topic.replyCount)}
          </span>
        </div>
      </div>
      <div role="cell" className="hidden justify-center lg:flex">
        <AvatarStack users={topic.participants} large />
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
}

function AvatarStack({
  users,
  large = false,
}: {
  users: TopicPayload["participants"];
  large?: boolean;
}) {
  const size = large ? "size-8 -ml-3" : "size-6 -ml-2";
  return (
    <div className="flex pl-3">
      {users.map((user) => (
        <UserCardPopover key={user.id} user={user}>
          <GooseLink
            href={`/u/${user.id}`}
            title={user.username}
            className={`${size} first:ml-0 rounded-full ring-2 ring-background transition-transform duration-150 hover:z-10 hover:scale-110 data-[state=open]:z-10 data-[state=open]:scale-110`}
          >
            <Avatar className="size-full">
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
