import { useState, type ReactNode } from "react";
import type { TopicPayload, UserProfileProps } from "@gooseforum/client";
import { formatCompactNumber } from "@gooseforum/client";
import {
  Award,
  Bird,
  CalendarDays,
  FileText,
  List,
  Radio,
  Settings,
  UserPlus,
  UserRound,
  MessageSquare,
} from "lucide-react";
import { siBilibili, siGithub, siSinaweibo, siX, siZhihu } from "simple-icons";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@gooseforum/ui/components/tooltip";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { ProfileAvatar } from "../users/profile-avatar";
import { ProfileBadge } from "../users/profile-badge";
import { UserActivity, UserActivityItem } from "../users/user-activity";

const socialProfiles = {
  github: { label: "GitHub", path: siGithub.path },
  twitter: { label: "X", path: siX.path },
  linkedIn: {
    label: "LinkedIn",
    path: "M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227V1.729C24 .774 23.2 0 22.222 0z",
  },
  weibo: { label: "Weibo", path: siSinaweibo.path },
  bilibili: { label: "Bilibili", path: siBilibili.path },
  zhihu: { label: "Zhihu", path: siZhihu.path },
} as const;

export function UserProfilePageView({ page }: { page: UserProfileProps }) {
  const { t } = useTranslation("user");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [following, setFollowing] = useState(page.user.isFollowing);
  const [followLoading, setFollowLoading] = useState(false);
  const [followError, setFollowError] = useState("");
  const displayName = page.user.nickname || page.user.username;
  const cover = safeUrl(page.user.profileCoverUrl, true);

  async function toggleFollow() {
    if (!page.canFollow || followLoading) return;
    setFollowLoading(true);
    setFollowError("");
    try {
      await runtime.api.users.follow(page.user.userId, following);
      setFollowing((current) => !current);
    } catch (reason) {
      setFollowError(
        serverError(reason, t("followFailed")),
      );
    } finally {
      setFollowLoading(false);
    }
  }

  return (
    <article className="pb-12">
      <section className="site-panel overflow-hidden rounded-xl border bg-background">
        <div
          data-slot="profile-cover"
          className="h-20 border-b bg-muted bg-cover bg-center lg:h-24"
          style={{
            backgroundImage: cover
              ? `url(${JSON.stringify(cover)}), ${defaultCover}`
              : defaultCover,
          }}
        />
        <div className="px-4 pb-4 lg:px-5">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div className="flex min-w-0 flex-1 gap-4">
              <ProfileAvatar
                src={page.user.avatarUrl}
                name={page.user.username}
                badge={page.user.wornBadge}
                className="-mt-9 size-24 lg:-mt-10 lg:size-28"
              />
              <div className="min-w-0 flex-1 pt-3">
                <div className="flex min-w-0 flex-wrap items-center gap-2">
                  <h1 className="truncate text-2xl font-bold leading-tight">
                    {displayName}
                  </h1>
                  {page.user.isAdmin ? (
                    <Badge variant="secondary" className="text-warning">
                      Admin
                    </Badge>
                  ) : null}
                  {page.user.isOnline ? (
                    <Badge variant="secondary" className="text-success">
                      <Radio data-icon="inline-start" />
                      {t("online")}
                    </Badge>
                  ) : null}
                </div>
                <div className="mt-1 flex min-w-0 flex-wrap items-center gap-1.5">
                  <p className="truncate text-sm font-medium text-muted-foreground">
                    @{page.user.username}
                  </p>
                  {page.isOwnProfile ? (
                    <Button asChild variant="ghost" size="xs">
                      <GooseLink href={page.settingsUrl}>
                        <Settings data-icon="inline-start" />
                        {t("editProfile")}
                      </GooseLink>
                    </Button>
                  ) : null}
                </div>
                <p className="mt-2 text-sm leading-relaxed text-foreground/75">
                  {page.user.bio || page.user.signature || t("emptyBio")}
                </p>
              </div>
            </div>
            {!page.isOwnProfile && (page.canMessage || page.canFollow) ? (
              <div className="flex shrink-0 flex-wrap items-center gap-2">
                {page.canMessage ? (
                  <Button asChild variant="outline">
                    <GooseLink href={page.messageUrl}>
                      <MessageSquare data-icon="inline-start" />
                      {t("message")}
                    </GooseLink>
                  </Button>
                ) : null}
                {page.canFollow ? (
                  <Button
                    type="button"
                    variant={following ? "secondary" : "default"}
                    disabled={followLoading}
                    onClick={() => void toggleFollow()}
                  >
                    <UserPlus data-icon="inline-start" />
                    {followLoading
                      ? t("loading")
                      : following
                        ? t("following")
                        : t("follow")}
                  </Button>
                ) : null}
              </div>
            ) : null}
          </div>
          {followError ? (
            <p className="mt-3 text-sm text-destructive">{followError}</p>
          ) : null}
          <div className="mt-3 flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-muted-foreground">
              <span className="inline-flex items-center gap-1.5">
                <CalendarDays className="size-3.5" />
                {t("joinedAt", {
                  date: formatDate(page.user.createdAt, runtime.locale),
                })}
              </span>
              {page.user.lastActiveTime ? (
                <span>
                  {t("lastActive", {
                    time: relativeTime(
                      page.user.lastActiveTime,
                      runtime.locale,
                      t("home:justNow"),
                    ),
                  })}
                </span>
              ) : null}
            </div>
            <ProfileLinks user={page.user} />
          </div>
        </div>

        <nav
          className="grid grid-cols-3 border-y"
          aria-label={t("profileNavigation")}
        >
          {page.tabs.map((tab) => {
            const Icon =
              tab.key === "summary"
                ? UserRound
                : tab.key === "activity"
                  ? List
                  : Award;
            return (
              <GooseLink
                key={tab.key}
                href={tab.url}
                aria-current={tab.active ? "page" : undefined}
                className={cn(
                  "inline-flex h-11 min-w-0 items-center justify-center gap-2 px-2 text-sm font-semibold text-muted-foreground hover:text-foreground",
                  tab.active && "border-b-2 border-primary text-primary",
                )}
              >
                <Icon className="size-4 shrink-0" />
                {t(`tabs.${tab.key}`)}
              </GooseLink>
            );
          })}
        </nav>

        {page.section === "summary" ? (
          <ProfileSummary page={page} locale={runtime.locale} t={t} />
        ) : page.section === "activity" ? (
          <UserActivity page={page} />
        ) : (
          <BadgeDirectory page={page} t={t} />
        )}
      </section>
    </article>
  );
}

const defaultCover =
  "linear-gradient(135deg, var(--muted) 0%, var(--gf-color-info-content) 52%, var(--muted) 100%)";

function ProfileSummary({
  page,
  locale,
  t,
}: {
  page: UserProfileProps;
  locale: string;
  t: ReturnType<typeof useTranslation>["t"];
}) {
  const stats = [
    ["reputation", page.user.prestige],
    ["topics", page.user.topicCount],
    ["replies", page.user.replyCount],
    ["likesReceived", page.user.likeReceivedCount],
    ["likesGiven", page.user.likeGivenCount],
    ["followers", page.user.followerCount],
    ["following", page.user.followingCount],
    ["bookmarks", page.user.collectionCount],
  ] as const;
  return (
    <div className="p-4">
      <section className="grid grid-cols-4 gap-y-4 border-b pb-4 lg:grid-cols-8">
        {stats.map(([key, value], index) => (
          <div key={key} className="min-w-0 text-center">
            <div
              className={cn(
                "text-base font-bold tabular-nums lg:text-lg",
                index === 0 && "text-primary",
              )}
            >
              {formatCompactNumber(value)}
            </div>
            <div
              className={cn(
                "mt-0.5 truncate text-[11px] font-medium text-muted-foreground lg:text-xs",
                index === 0 && "text-primary/80",
              )}
            >
              {t(`stats.${key}`)}
            </div>
          </div>
        ))}
      </section>
      <div className="grid gap-5 pt-4 lg:grid-cols-[minmax(0,1fr)_300px]">
        <section className="min-w-0">
          <h2 className="mb-2 text-sm font-semibold text-foreground/75">
            {t("summarySections.recentTopics")}
          </h2>
          <div className="divide-y">
            {page.topics.map((topic) => (
              <CompactTopic key={topic.id} topic={topic} locale={locale} />
            ))}
          </div>
          {!page.topics.length ? (
            <SimpleEmpty icon={FileText} title={t("emptyTopics")} />
          ) : null}
        </section>
        <aside className="flex min-w-0 flex-col gap-5">
          {page.badges.length ? (
            <section className="border-b pb-5 lg:border-b-0">
              <h2 className="mb-3 text-sm font-semibold text-foreground/75">
                {t("summarySections.recentBadges")}
              </h2>
              <div className="flex flex-wrap gap-x-3 gap-y-2">
                {page.badges.slice(0, 8).map((badge) => (
                  <div
                    key={badge.code}
                    className="group flex w-16 flex-col items-center gap-1"
                  >
                    <ProfileBadge badge={badge} compact />
                    <span className="w-full truncate text-center text-[10px] font-semibold text-foreground/75">
                      {badge.name}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ) : null}
          {page.activities.length ? (
            <section>
              <h2 className="mb-2 text-sm font-semibold text-foreground/75">
                {t("summarySections.recentActivity")}
              </h2>
              <div data-slot="recent-user-activity" className="flex flex-col gap-2">
                {page.activities.map((activity) => (
                  <UserActivityItem
                    key={activity.id}
                    activity={activity}
                    locale={locale}
                    t={t}
                  />
                ))}
              </div>
            </section>
          ) : null}
        </aside>
      </div>
    </div>
  );
}

function CompactTopic({
  topic,
  locale,
}: {
  topic: TopicPayload;
  locale: string;
}) {
  return (
    <GooseLink
      href={topic.url}
      className="grid gap-2 py-3 lg:grid-cols-[minmax(0,1fr)_72px_88px] lg:items-center"
    >
      <span className="min-w-0">
        <span className="flex min-w-0 flex-wrap items-center gap-2">
          <span className="truncate text-[15px] font-semibold">
            {topic.title}
          </span>
          {topic.categories.slice(0, 2).map((category) => (
            <Badge
              key={category.id}
              variant="secondary"
              className="gap-1 font-normal"
            >
              <span
                className="size-1.5 rounded-full"
                style={{ backgroundColor: category.color }}
              />
              {category.name}
            </Badge>
          ))}
        </span>
        <span className="mt-1 block truncate text-sm text-muted-foreground">
          {topic.description || "(｀・ω・´)"}
        </span>
      </span>
      <span className="hidden text-center text-sm font-semibold tabular-nums text-foreground/75 lg:block">
        {formatCompactNumber(topic.replyCount)}
      </span>
      <span className="hidden text-right text-xs font-medium text-muted-foreground lg:block">
        {relativeTime(topic.lastUpdateTime, locale)}
      </span>
    </GooseLink>
  );
}

function BadgeDirectory({
  page,
  t,
}: {
  page: UserProfileProps;
  t: ReturnType<typeof useTranslation>["t"];
}) {
  return page.badges.length ? (
    <div className="grid lg:grid-cols-4 lg:gap-3 lg:p-4">
      {page.badges.map((badge) => (
        <div
          key={badge.code}
          className="flex min-w-0 items-center gap-3 border-b p-3 lg:rounded-lg lg:border"
          title={badge.description}
        >
          <ProfileBadge badge={badge} />
          <span className="min-w-0">
            <span className="block truncate text-sm font-semibold">
              {badge.name}
            </span>
            <span className="mt-0.5 block truncate text-xs text-muted-foreground">
              {badge.description}
            </span>
          </span>
        </div>
      ))}
    </div>
  ) : (
    <SimpleEmpty icon={Award} title={t("emptyData")} />
  );
}

function ProfileLinks({ user }: { user: UserProfileProps["user"] }) {
  const website = safeUrl(user.website);
  const links = Object.entries(socialProfiles).flatMap(([key, icon]) => {
    const href = safeUrl(user.externalInformation?.[key]?.link);
    return href ? [{ key, href, ...icon }] : [];
  });
  if (!website && !links.length) return null;
  return (
    <TooltipProvider delayDuration={150}>
      <div className="flex flex-wrap items-center gap-0.5 lg:justify-end">
        {website ? (
          <IconLink href={website} label={user.websiteName || user.website}>
            <Bird className="size-5" />
          </IconLink>
        ) : null}
        {links.map((item) => (
          <IconLink key={item.key} href={item.href} label={item.label}>
            <svg
              className="size-4 fill-current"
              role="img"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path d={item.path} />
            </svg>
          </IconLink>
        ))}
      </div>
    </TooltipProvider>
  );
}

function IconLink({
  href,
  label,
  children,
}: {
  href: string;
  label: string;
  children: ReactNode;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <a
          href={href}
          target="_blank"
          rel="noopener noreferrer ugc"
          aria-label={label}
          className="inline-flex size-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-primary"
        >
          {children}
        </a>
      </TooltipTrigger>
      <TooltipContent side="top" sideOffset={8}>
        {label}
      </TooltipContent>
    </Tooltip>
  );
}

function SimpleEmpty({
  icon: Icon,
  title,
}: {
  icon: typeof FileText;
  title: string;
}) {
  return (
    <Empty className="min-h-40 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Icon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
      </EmptyHeader>
    </Empty>
  );
}

function safeUrl(value?: string, allowRelative = false) {
  const input = value?.trim();
  if (!input) return "";
  try {
    const parsed = new URL(input, "http://gooseforum.local");
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return "";
    return allowRelative && input.startsWith("/") ? input : parsed.toString();
  } catch {
    return "";
  }
}

function formatDate(value: string, locale: string) {
  const date = new Date(value.includes("T") ? value : value.replace(" ", "T"));
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleDateString(locale, { dateStyle: "medium" });
}

function relativeTime(value: string, locale: string, justNow?: string) {
  const date = new Date(value.includes("T") ? value : value.replace(" ", "T"));
  if (Number.isNaN(date.getTime())) return value;
  const seconds = Math.round((date.getTime() - Date.now()) / 1000);
  const formatter = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });
  if (Math.abs(date.getTime() - Date.now()) < 60_000)
    return justNow || formatter.format(seconds, "second");
  if (Math.abs(seconds) < 3600)
    return formatter.format(Math.round(seconds / 60), "minute");
  if (Math.abs(seconds) < 86400)
    return formatter.format(Math.round(seconds / 3600), "hour");
  if (Math.abs(seconds) < 604800)
    return formatter.format(Math.round(seconds / 86400), "day");
  return formatDate(value, locale);
}
