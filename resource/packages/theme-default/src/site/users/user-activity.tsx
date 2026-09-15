import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type RefObject,
} from "react";
import type {
  PagePayload,
  UserActivityPayload,
  UserProfileProps,
} from "@gooseforum/client";
import {
  FileText,
  Heart,
  List,
  MessageCircle,
  PenLine,
  UserPlus,
  UserRound,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink, useGooseLocale, useGoosePageFetcher } from "@gooseforum/runtime";
import { TopicListFooter, TopicTable } from "../topics/topic-list";

export function UserActivity({ page }: { page: UserProfileProps }) {
  const { t } = useTranslation("user");
  const { t: topicT } = useTranslation("home");
  const locale = useGooseLocale();
  const fetchPage = useGoosePageFetcher();
  const [topics, setTopics] = useState(page.topics);
  const [activities, setActivities] = useState(page.activities);
  const [likes, setLikes] = useState(page.likes);
  const [pagination, setPagination] = useState(page.pagination);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const sentinel = useRef<HTMLDivElement>(null);

  const loadMore = useCallback(async () => {
    if (
      loading ||
      !pagination.hasNext ||
      !pagination.nextUrl ||
      !fetchPage
    )
      return;
    setLoading(true);
    setError("");
    try {
      const next = (await fetchPage(
        pagination.nextUrl,
      )) as PagePayload<UserProfileProps>;
      if (page.activityTab === "topics") {
        setTopics((current) => mergeUnique(current, next.props.topics));
      } else if (page.activityTab === "likes") {
        setLikes((current) => mergeUnique(current, next.props.likes));
      } else {
        setActivities((current) => mergeUnique(current, next.props.activities));
      }
      setPagination(next.props.pagination);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [loading, page.activityTab, pagination, fetchPage, t]);

  useEffect(() => {
    if (!sentinel.current || !("IntersectionObserver" in window)) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) void loadMore();
      },
      { rootMargin: "480px 0px" },
    );
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [loadMore]);

  const connections =
    page.activityTab === "following" ? page.following : page.followers;
  return (
    <div>
      <nav
        className="grid grid-cols-5 border-b"
        aria-label={t("activityNavigation")}
      >
        {page.activityTabs.map((tab) => (
          <ProfileTabLink
            key={tab.key}
            tab={tab}
            label={activityTabLabel(tab.key, t)}
          />
        ))}
      </nav>

      {page.activityTab === "topics" ? (
        <>
          <TopicTable topics={topics} t={topicT} />
          {!topics.length ? (
            <ProfileEmpty icon={FileText} title={t("emptyTopics")} />
          ) : null}
          <FeedFooter
            visible={topics.length > 0 || pagination.hasNext}
            pagination={pagination}
            loading={loading}
            error={error}
            topicT={topicT}
            loadMore={loadMore}
            sentinel={sentinel}
          />
        </>
      ) : page.activityTab === "timeline" ? (
        <>
          <div className="flex flex-col gap-0 lg:gap-3 lg:p-4">
            {activities.map((activity) => (
              <UserActivityItem
                key={activity.id}
                activity={activity}
                locale={locale}
                t={t}
              />
            ))}
            {!activities.length ? (
              <ProfileEmpty icon={MessageCircle} title={t("emptyActivity")} />
            ) : null}
          </div>
          <FeedFooter
            visible={activities.length > 0 || pagination.hasNext}
            pagination={pagination}
            loading={loading}
            error={error}
            topicT={topicT}
            loadMore={loadMore}
            sentinel={sentinel}
          />
        </>
      ) : page.activityTab === "likes" ? (
        <>
          <div className="flex flex-col gap-0 lg:gap-3 lg:p-4">
            {likes.map((like) => (
              <GooseLink
                key={like.id}
                href={like.url}
                className="flex min-w-0 gap-3 border-b p-4 lg:rounded-lg lg:border lg:p-3 transition hover:border-primary/20 hover:bg-muted/40"
              >
                <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
                  <Heart className="size-4" />
                </span>
                <span className="min-w-0">
                  <span className="block text-xs font-medium text-muted-foreground">
                    {t("activity.like")}
                  </span>
                  <span className="mt-0.5 block truncate text-sm font-semibold">
                    {like.title}
                  </span>
                  <time className="mt-1 block text-xs text-muted-foreground">
                    {formatDateTime(like.likedAt, locale)}
                  </time>
                </span>
              </GooseLink>
            ))}
            {!likes.length ? (
              <ProfileEmpty icon={Heart} title={t("emptyData")} />
            ) : null}
          </div>
          <FeedFooter
            visible={likes.length > 0 || pagination.hasNext}
            pagination={pagination}
            loading={loading}
            error={error}
            topicT={topicT}
            loadMore={loadMore}
            sentinel={sentinel}
          />
        </>
      ) : connections.length ? (
        <div className="grid gap-0 lg:gap-3 lg:p-4 lg:grid-cols-2 xl:grid-cols-3">
          {connections.map((connection) => (
            <GooseLink
              key={connection.id}
              href={connection.url}
              className="flex min-w-0 gap-3 border-b p-4 lg:rounded-lg lg:border lg:p-3 transition hover:border-primary/20 hover:bg-muted/40"
            >
              <Avatar className="size-10">
                <AvatarImage
                  src={connection.avatarUrl}
                  alt={connection.username}
                />
                <AvatarFallback>
                  {connection.username.slice(0, 1).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="min-w-0">
                <span className="block truncate text-sm font-semibold">
                  {connection.nickname || connection.username}
                </span>
                <span className="block truncate text-xs text-muted-foreground">
                  @{connection.username}
                </span>
                <span className="mt-1 block truncate text-xs text-muted-foreground">
                  {connection.bio || t("noBio")}
                </span>
              </span>
            </GooseLink>
          ))}
        </div>
      ) : (
        <ProfileEmpty icon={UserPlus} title={t("emptyData")} />
      )}
    </div>
  );
}

function ProfileTabLink({
  tab,
  label,
}: {
  tab: UserProfileProps["activityTabs"][number];
  label: string;
}) {
  const Icon =
    tab.key === "timeline"
      ? List
      : tab.key === "topics"
        ? FileText
        : tab.key === "likes"
          ? Heart
          : tab.key === "following"
            ? UserPlus
            : UserRound;
  return (
    <GooseLink
      href={tab.url}
      aria-current={tab.active ? "page" : undefined}
      className={
        tab.active
          ? "inline-flex h-10 min-w-0 items-center justify-center gap-2 border-b-2 border-primary px-2 text-sm font-medium text-primary"
          : "inline-flex h-10 min-w-0 items-center justify-center gap-2 px-2 text-sm font-medium text-muted-foreground hover:text-foreground"
      }
    >
      <Icon className="hidden size-4 shrink-0 lg:block" />
      <span className="truncate">{label}</span>
    </GooseLink>
  );
}

export function UserActivityItem({
  activity,
  locale,
  t,
}: {
  activity: UserActivityPayload;
  locale: string;
  t: (key: string) => string;
}) {
  const Icon =
    activity.action === 2
      ? PenLine
      : activity.action === 3
        ? Heart
        : activity.action === 4
          ? UserPlus
          : activity.action === 5
            ? MessageCircle
            : FileText;
  const label = activityLabel(activity, t);
  return (
    <GooseLink
      href={activity.url || "#"}
      className="flex gap-3 border-b p-4 lg:rounded-lg lg:border lg:p-3 transition hover:border-primary/20 hover:bg-muted/40"
    >
      <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
        <Icon className="size-4" />
      </span>
      <span className="min-w-0">
        <span className="block truncate text-sm font-semibold">
          {activity.contentPreview
            ? `${label}: ${activity.contentPreview}`
            : label}
        </span>
        <time className="mt-1 block text-xs text-muted-foreground">
          {formatDateTime(activity.createdAt, locale)}
        </time>
      </span>
    </GooseLink>
  );
}

function FeedFooter({
  visible,
  pagination,
  loading,
  error,
  topicT,
  loadMore,
  sentinel,
}: {
  visible: boolean;
  pagination: UserProfileProps["pagination"];
  loading: boolean;
  error: string;
  topicT: ReturnType<typeof useTranslation>["t"];
  loadMore(): Promise<void>;
  sentinel: RefObject<HTMLDivElement | null>;
}) {
  if (!visible) return null;
  return (
    <div ref={sentinel}>
      <TopicListFooter
        mode="waterfall"
        pagination={pagination}
        loading={loading}
        error={error}
        t={topicT}
        onLoad={loadMore}
      />
    </div>
  );
}

function ProfileEmpty({
  icon: Icon,
  title,
}: {
  icon: typeof FileText;
  title: string;
}) {
  return (
    <Empty className="min-h-44 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Icon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription className="sr-only">{title}</EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}

function activityTabLabel(key: string, t: (key: string) => string) {
  return t(`tabs.${key}`);
}

function activityLabel(
  activity: UserActivityPayload,
  t: (key: string) => string,
) {
  if (activity.action === 1 || activity.label === "signup")
    return t("activity.signup");
  if (activity.action === 2 || activity.label === "post")
    return t("activity.post");
  if (activity.action === 3 || activity.label === "like")
    return t("activity.like");
  if (activity.action === 4 || activity.label === "follow")
    return t("activity.follow");
  if (activity.action === 5 || activity.label === "comment")
    return t("activity.comment");
  return activity.label || t("activity.default");
}

function mergeUnique<T extends { id: number }>(current: T[], incoming: T[]) {
  const seen = new Set(current.map((item) => item.id));
  const unique = incoming.filter((item) => {
    if (seen.has(item.id)) return false;
    seen.add(item.id);
    return true;
  });
  return [...current, ...unique];
}

function formatDateTime(value: string, locale: string) {
  const date = new Date(value.includes("T") ? value : value.replace(" ", "T"));
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });
}
