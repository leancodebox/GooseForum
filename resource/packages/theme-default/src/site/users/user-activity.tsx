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
import { formatDate } from "../date-format";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, useGooseLocale, useGoosePageFetcher, useLatestRequest } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { TopicListFooter, TopicTable } from "../topics/topic-list";
import { appendUniqueById } from "../collection-utils";
import { useIntersectionSentinel } from "../use-intersection-sentinel";

export function UserActivity({ page }: { page: UserProfileProps }) {
  const { t } = useTranslation("user");
  const { t: topicT } = useTranslation("home");
  const serverError = useServerErrorMessage();
  const locale = useGooseLocale();
  const fetchPage = useGoosePageFetcher();
  const [topics, setTopics] = useState(page.topics);
  const [activities, setActivities] = useState(page.activities);
  const [likes, setLikes] = useState(page.likes);
  const [pagination, setPagination] = useState(page.pagination);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const sentinel = useRef<HTMLDivElement>(null);
  const loadingRef = useRef(loading);
  const paginationRef = useRef(pagination);
  const activityTabRef = useRef(page.activityTab);
  const beginRequest = useLatestRequest();

  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);

  useEffect(() => {
    paginationRef.current = pagination;
  }, [pagination]);

  useEffect(() => {
    if (activityTabRef.current === page.activityTab) return;
    activityTabRef.current = page.activityTab;
    beginRequest();
    loadingRef.current = false;
    setTopics(page.topics);
    setActivities(page.activities);
    setLikes(page.likes);
    setPagination(page.pagination);
    setLoading(false);
    setError("");
  }, [beginRequest, page.activityTab, page.activities, page.likes, page.pagination, page.topics]);

  const loadMore = useCallback(async () => {
    const currentPagination = paginationRef.current;
    if (
      loadingRef.current ||
      !currentPagination.hasNext ||
      !currentPagination.nextUrl ||
      !fetchPage
    )
      return;
    const isCurrent = beginRequest();
    const currentTab = page.activityTab;
    loadingRef.current = true;
    setLoading(true);
    setError("");
    try {
      const next = (await fetchPage(
        currentPagination.nextUrl,
      )) as PagePayload<UserProfileProps>;
      if (!isCurrent() || currentTab !== activityTabRef.current) return;
      if (currentTab === "topics") {
        setTopics((current) => appendUniqueById(current, next.props.topics));
      } else if (currentTab === "likes") {
        setLikes((current) => appendUniqueById(current, next.props.likes));
      } else {
        setActivities((current) => appendUniqueById(current, next.props.activities));
      }
      setPagination(next.props.pagination);
    } catch (reason) {
      if (isCurrent() && currentTab === activityTabRef.current) {
        setError(serverError(reason, t("loadFailed")));
      }
    } finally {
      loadingRef.current = false;
      if (isCurrent() && currentTab === activityTabRef.current) {
        setLoading(false);
      }
    }
  }, [beginRequest, page.activityTab, fetchPage, serverError, t]);

  useIntersectionSentinel(sentinel, () => void loadMore());

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
                    {formatDate(like.likedAt, locale)}
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
      className={cn(
        "inline-flex h-10 min-w-0 items-center justify-center gap-2 px-2 text-sm font-medium",
        tab.active
          ? "border-b-2 border-primary text-primary"
          : "text-muted-foreground hover:text-foreground",
      )}
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
          {formatDate(activity.createdAt, locale)}
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
