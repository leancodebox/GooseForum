import { useCallback, useEffect, useState, type ComponentType } from "react";
import type {
  ModerationLogItem,
  ModerationPageProps,
  ModerationReportItem,
  TopicPayload,
} from "@gooseforum/client";
import {
  Ban,
  CircleAlert,
  Flag,
  History,
  RotateCcw,
  Scale,
  XCircle,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "../../components/ui/alert";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { Spinner } from "../../components/ui/spinner";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "../../components/ui/tabs";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { PageHeader } from "../layout/page-header";
import { UserCardPopover } from "../users/user-card-popover";

type ConsoleTab = "reports" | "ban" | "logs" | "guidance";
type ReportStatus = "open" | "closed";
const consoleTabs: Array<{ key: ConsoleTab; icon: ComponentType }> = [
  { key: "reports", icon: Flag },
  { key: "ban", icon: Ban },
  { key: "logs", icon: History },
  { key: "guidance", icon: Scale },
];

export function ModerationPageView({ page }: { page: ModerationPageProps }) {
  const { t } = useTranslation("moderation");
  const runtime = useGooseRuntime();
  const [tab, setTab] = useState<ConsoleTab>("reports");
  const [topics, setTopics] = useState(page.topics);
  const [topicBusy, setTopicBusy] = useState<Set<number>>(() => new Set());
  const [topicError, setTopicError] = useState("");
  const [reportStatus, setReportStatus] = useState<ReportStatus>("open");
  const [reports, setReports] = useState<ModerationReportItem[]>([]);
  const [reportCursor, setReportCursor] = useState(0);
  const [reportMore, setReportMore] = useState(true);
  const [reportLoading, setReportLoading] = useState(false);
  const [reportLoaded, setReportLoaded] = useState(false);
  const [reportError, setReportError] = useState("");
  const [reportBusy, setReportBusy] = useState<Set<number>>(() => new Set());
  const [logs, setLogs] = useState<ModerationLogItem[]>([]);
  const [logCursor, setLogCursor] = useState(0);
  const [logMore, setLogMore] = useState(true);
  const [logLoading, setLogLoading] = useState(false);
  const [logLoaded, setLogLoaded] = useState(false);
  const [logError, setLogError] = useState("");

  useEffect(() => {
    setTopics(page.topics);
    setTopicError("");
  }, [page]);
  const loadReports = useCallback(
    async (reset = false, status = reportStatus) => {
      if (reportLoading) return;
      setReportLoading(true);
      setReportError("");
      try {
        const result = await runtime.api.moderation.reports(
          reset ? 0 : reportCursor,
          20,
          status,
        );
        setReports((current) =>
          reset ? result.items : merge(current, result.items),
        );
        setReportCursor(result.nextCursor);
        setReportMore(result.hasNext);
        setReportLoaded(true);
      } catch (reason) {
        setReportError(
          reason instanceof Error ? reason.message : t("reports.loadFailed"),
        );
      } finally {
        setReportLoading(false);
      }
    },
    [reportCursor, reportLoading, reportStatus, runtime.api.moderation, t],
  );
  const loadLogs = useCallback(
    async (reset = false) => {
      if (logLoading) return;
      setLogLoading(true);
      setLogError("");
      try {
        const result = await runtime.api.moderation.logs(
          reset ? 0 : logCursor,
          20,
        );
        setLogs((current) =>
          reset ? result.items : merge(current, result.items),
        );
        setLogCursor(result.nextCursor);
        setLogMore(result.hasNext);
        setLogLoaded(true);
      } catch (reason) {
        setLogError(
          reason instanceof Error ? reason.message : t("logs.loadFailed"),
        );
      } finally {
        setLogLoading(false);
      }
    },
    [logCursor, logLoading, runtime.api.moderation, t],
  );
  useEffect(() => {
    if (tab === "reports" && !reportLoaded) void loadReports(true);
    if (tab === "logs" && !logLoaded) void loadLogs(true);
  }, [loadLogs, loadReports, logLoaded, reportLoaded, tab]);

  function changeReportStatus(status: ReportStatus) {
    if (status === reportStatus) return;
    setReportStatus(status);
    setReports([]);
    setReportCursor(0);
    setReportMore(true);
    setReportLoaded(false);
    void loadReports(true, status);
  }
  async function handleReport(
    item: ModerationReportItem,
    action: "ban" | "reject",
  ) {
    if (reportBusy.has(item.id)) return;
    setReportBusy((current) => withId(current, item.id));
    setReportError("");
    try {
      if (action === "ban")
        await (item.targetType === "topic"
          ? runtime.api.moderation.setTopicStatus(item.targetId, "ban")
          : runtime.api.moderation.setPostStatus(item.targetId, "ban"));
      await runtime.api.moderation.setReportStatus(item.id, action);
      setReports((items) => items.filter((entry) => entry.id !== item.id));
      if (item.targetType === "topic")
        setTopics((items) =>
          items.filter((entry) => entry.id !== item.targetId),
        );
      setLogLoaded(false);
    } catch (reason) {
      setReportError(
        reason instanceof Error ? reason.message : t("reports.actionFailed"),
      );
    } finally {
      setReportBusy((current) => withoutId(current, item.id));
    }
  }
  async function restoreTopic(topic: TopicPayload) {
    if (topicBusy.has(topic.id)) return;
    setTopicBusy((current) => withId(current, topic.id));
    setTopicError("");
    try {
      await runtime.api.moderation.setTopicStatus(topic.id, "unban");
      setTopics((items) => items.filter((entry) => entry.id !== topic.id));
      setLogLoaded(false);
    } catch (reason) {
      setTopicError(
        reason instanceof Error ? reason.message : t("blocked.loadFailed"),
      );
    } finally {
      setTopicBusy((current) => withoutId(current, topic.id));
    }
  }

  return (
    <main className="min-w-0 pb-8">
      <PageHeader compact title={t("title")} description={t("description")} />
      <Tabs
        value={tab}
        onValueChange={(value) => setTab(value as ConsoleTab)}
        className="mb-4"
      >
        <div
          data-slot="moderation-tabs-frame"
          className="w-fit max-w-full rounded-xl bg-muted p-1 ring-1 ring-border/50"
        >
          <TabsList className="h-auto max-w-full justify-start gap-0.5 overflow-x-auto rounded-lg bg-transparent p-0 group-data-horizontal/tabs:h-auto">
            {consoleTabs.map(({ key, icon: Icon }) => (
              <TabsTrigger
                key={key}
                value={key}
                className="h-8 flex-none px-3"
              >
                <Icon />
                {t(`tabs.${key}`)}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        <TabsContent value="reports">
          <ReportsPanel
            status={reportStatus}
            items={reports}
            loading={reportLoading}
            loaded={reportLoaded}
            hasMore={reportMore}
            error={reportError}
            busy={reportBusy}
            t={t}
            locale={runtime.locale}
            onStatus={changeReportStatus}
            onLoad={() => void loadReports()}
            onAction={(item, action) => void handleReport(item, action)}
          />
        </TabsContent>
        <TabsContent value="ban">
          <BlockedPanel
            page={page}
            topics={topics}
            busy={topicBusy}
            error={topicError}
            t={t}
            onRestore={(topic) => void restoreTopic(topic)}
          />
        </TabsContent>
        <TabsContent value="logs">
          <LogsPanel
            items={logs}
            loading={logLoading}
            loaded={logLoaded}
            hasMore={logMore}
            error={logError}
            t={t}
            locale={runtime.locale}
            onLoad={() => void loadLogs()}
          />
        </TabsContent>
        <TabsContent value="guidance">
          <Guidance t={t} />
        </TabsContent>
      </Tabs>
    </main>
  );
}

type Translate = ReturnType<typeof useTranslation>["t"];
function ReportsPanel({
  status,
  items,
  loading,
  loaded,
  hasMore,
  error,
  busy,
  t,
  locale,
  onStatus,
  onLoad,
  onAction,
}: {
  status: ReportStatus;
  items: ModerationReportItem[];
  loading: boolean;
  loaded: boolean;
  hasMore: boolean;
  error: string;
  busy: Set<number>;
  t: Translate;
  locale: string;
  onStatus(value: ReportStatus): void;
  onLoad(): void;
  onAction(item: ModerationReportItem, action: "ban" | "reject"): void;
}) {
  return (
    <section className="flex flex-col gap-3">
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="overflow-hidden rounded-xl border bg-background">
        <Tabs
          value={status}
          onValueChange={(value) => onStatus(value as ReportStatus)}
          className="gap-0"
        >
          <TabsList className="h-auto w-full justify-start gap-1 rounded-none border-b bg-muted/50 p-2 group-data-horizontal/tabs:h-auto">
            <TabsTrigger
              value="open"
              className="h-8 flex-none rounded-md px-3 font-semibold"
            >
              {t("reports.statusTabs.open")}
            </TabsTrigger>
            <TabsTrigger
              value="closed"
              className="h-8 flex-none rounded-md px-3 font-semibold"
            >
              {t("reports.statusTabs.closed")}
            </TabsTrigger>
          </TabsList>
        </Tabs>
        {items.length ? (
          <div className="divide-y">
            {items.map((item) => (
              <article
                key={item.id}
                className="grid grid-cols-[28px_minmax(0,1fr)] gap-3 px-3 py-2.5 hover:bg-muted/40 lg:grid-cols-[28px_minmax(0,1fr)_150px_180px_auto] lg:items-center"
              >
                <span className="flex size-7 items-center justify-center rounded bg-muted text-amber-600">
                  <Flag />
                </span>
                <div className="min-w-0">
                  <div className="flex min-w-0 flex-wrap items-center gap-1.5">
                    <span className="text-xs text-muted-foreground">
                      {t(`reports.targetTypes.${item.targetType}`)}
                    </span>
                    <GooseLink
                      href={item.targetUrl}
                      className="min-w-0 truncate font-medium text-primary"
                    >
                      {item.title}
                    </GooseLink>
                    {item.categories.map((category) => (
                      <span
                        key={category.id}
                        className="inline-flex items-center gap-1 text-xs text-muted-foreground"
                      >
                        <span
                          className="size-2 rounded-sm"
                          style={{ backgroundColor: category.color }}
                        />
                        {category.name}
                      </span>
                    ))}
                  </div>
                  <p className="line-clamp-1 text-sm text-muted-foreground">
                    {item.excerpt || item.note || t("reports.noExcerpt")}
                  </p>
                  <time className="text-xs text-muted-foreground lg:hidden">
                    {formatDate(item.createdAt, locale)}
                  </time>
                </div>
                <div className="hidden min-w-0 text-sm lg:block">
                  <p>
                    <span className="text-muted-foreground">
                      {t("reports.reason")}{" "}
                    </span>
                    {t(`reports.reasons.${item.reason}`, {
                      defaultValue: item.reason,
                    })}
                  </p>
                  {status === "closed" ? (
                    <p>
                      <span className="text-muted-foreground">
                        {t("reports.status")}{" "}
                      </span>
                      {resolution(item, t)}
                    </p>
                  ) : null}
                </div>
                <div className="hidden min-w-0 lg:block">
                  <Person user={item.reporter} label={t("reports.reporter")} />
                  {status === "closed" && item.handler.id ? (
                    <Person user={item.handler} label={t("reports.handler")} />
                  ) : null}
                </div>
                <div className="col-start-2 flex flex-wrap items-center gap-2 lg:col-start-auto lg:justify-end">
                  {status === "open" ? (
                    <>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={busy.has(item.id)}
                        onClick={() => onAction(item, "ban")}
                      >
                        <Ban data-icon="inline-start" />
                        {t("reports.block")}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={busy.has(item.id)}
                        onClick={() => onAction(item, "reject")}
                      >
                        <XCircle data-icon="inline-start" />
                        {t("reports.ignore")}
                      </Button>
                    </>
                  ) : (
                    <time className="text-xs text-muted-foreground">
                      {formatDate(item.handledAt || item.createdAt, locale)}
                    </time>
                  )}
                </div>
              </article>
            ))}
          </div>
        ) : (
          <PanelEmpty
            loading={loading}
            icon={Flag}
            title={t(loading ? "reports.loading" : "reports.emptyTitle")}
            description={loading ? "" : t("reports.emptyDescription")}
          />
        )}
        {loaded && (items.length || hasMore) ? (
          <footer className="border-t px-4 py-3 text-center">
            {hasMore ? (
              <Button
                variant="ghost"
                size="sm"
                disabled={loading}
                onClick={onLoad}
              >
                {loading ? <Spinner data-icon="inline-start" /> : null}
                {loading ? t("reports.loading") : t("reports.loadMore")}
              </Button>
            ) : (
              <span className="text-xs text-muted-foreground">
                {t("reports.noMore")}
              </span>
            )}
          </footer>
        ) : null}
      </div>
    </section>
  );
}
function BlockedPanel({
  page,
  topics,
  busy,
  error,
  t,
  onRestore,
}: {
  page: ModerationPageProps;
  topics: TopicPayload[];
  busy: Set<number>;
  error: string;
  t: Translate;
  onRestore(topic: TopicPayload): void;
}) {
  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-wrap gap-2">
        {page.categoryTabs.map((tab) => (
          <Button
            key={tab.key}
            asChild
            size="sm"
            variant={tab.active ? "default" : "ghost"}
            className={
              tab.active
                ? "shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            }
          >
            <GooseLink
              href={tab.url}
              aria-current={tab.active ? "page" : undefined}
            >
              {tab.label}
            </GooseLink>
          </Button>
        ))}
      </div>
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="overflow-hidden rounded-xl border bg-background">
        {topics.length ? (
          <div className="divide-y">
            {topics.map((topic) => (
              <article
                key={topic.id}
                className="flex items-center justify-between gap-4 px-4 py-3 hover:bg-muted/40"
              >
                <div className="min-w-0">
                  <GooseLink
                    href={topic.url}
                    className="block truncate font-semibold hover:text-primary"
                  >
                    {topic.title}
                  </GooseLink>
                  <p className="truncate text-sm text-muted-foreground">
                    {topic.description}
                  </p>
                </div>
                <Button
                  size="sm"
                  disabled={busy.has(topic.id)}
                  onClick={() => onRestore(topic)}
                >
                  {busy.has(topic.id) ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <RotateCcw data-icon="inline-start" />
                  )}
                  {t("blocked.restore")}
                </Button>
              </article>
            ))}
          </div>
        ) : (
          <PanelEmpty
            icon={Ban}
            title={t("blocked.emptyTitle")}
            description={t("blocked.emptyDescription")}
          />
        )}
        {page.pagination.hasNext ? (
          <footer className="border-t px-4 py-3 text-center">
            <Button asChild variant="outline" size="sm">
              <GooseLink href={page.pagination.nextUrl}>
                {t("blocked.next")}
              </GooseLink>
            </Button>
          </footer>
        ) : null}
      </div>
    </section>
  );
}
function LogsPanel({
  items,
  loading,
  loaded,
  hasMore,
  error,
  t,
  locale,
  onLoad,
}: {
  items: ModerationLogItem[];
  loading: boolean;
  loaded: boolean;
  hasMore: boolean;
  error: string;
  t: Translate;
  locale: string;
  onLoad(): void;
}) {
  return (
    <section className="flex flex-col gap-3">
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="overflow-hidden rounded-xl border bg-background">
        {items.length ? (
          <div className="divide-y">
            {items.map((item) => (
              <article
                key={item.id}
                className="grid grid-cols-[34px_minmax(0,1fr)] gap-3 px-3 py-2.5 hover:bg-muted/40 md:grid-cols-[34px_minmax(0,1fr)_150px]"
              >
                <span className="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
                  <History />
                </span>
                <div className="min-w-0">
                  <p className="truncate text-sm">
                    <strong>{item.actor.username}</strong>{" "}
                    <span className="text-muted-foreground">
                      {t(`logs.actions.${item.action}`, {
                        defaultValue: t("logs.actions.operation"),
                      })}
                    </span>{" "}
                    {item.subject.url ? (
                      <GooseLink
                        href={item.subject.url}
                        className="font-semibold text-primary"
                      >
                        {item.subject.title}
                      </GooseLink>
                    ) : (
                      <strong>{item.subject.title}</strong>
                    )}
                  </p>
                  {item.subject.excerpt ? (
                    <p className="truncate text-xs text-muted-foreground">
                      {item.subject.excerpt}
                    </p>
                  ) : null}
                </div>
                <time className="hidden text-right text-xs text-muted-foreground md:block">
                  {formatDate(item.createdAt, locale)}
                </time>
              </article>
            ))}
          </div>
        ) : (
          <PanelEmpty
            loading={loading}
            icon={History}
            title={t(loading ? "logs.loading" : "logs.emptyTitle")}
            description={loading ? "" : t("logs.emptyDescription")}
          />
        )}
        {loaded && (items.length || hasMore) ? (
          <footer className="border-t px-4 py-3 text-center">
            {hasMore ? (
              <Button
                variant="ghost"
                size="sm"
                disabled={loading}
                onClick={onLoad}
              >
                {loading ? <Spinner data-icon="inline-start" /> : null}
                {loading ? t("logs.loading") : t("logs.loadMore")}
              </Button>
            ) : (
              <span className="text-xs text-muted-foreground">
                {t("logs.noMore")}
              </span>
            )}
          </footer>
        ) : null}
      </div>
    </section>
  );
}
function Guidance({ t }: { t: Translate }) {
  return (
    <section className="px-4 sm:px-0">
      <div className="flex items-start gap-3 py-2.5">
        <CircleAlert className="mt-1 text-amber-600" />
        <p className="text-sm leading-6 text-muted-foreground">{t("notice")}</p>
      </div>
      <div className="flex flex-col gap-1">
        {(["rule", "context", "restraint"] as const).map((key) => (
          <article key={key} className="py-2">
            <h3 className="text-sm font-semibold">
              {t(`guidance.${key}.title`)}
            </h3>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">
              {t(`guidance.${key}.description`)}
            </p>
          </article>
        ))}
      </div>
    </section>
  );
}
function Person({
  user,
  label,
}: {
  user: ModerationReportItem["reporter"];
  label: string;
}) {
  return (
    <UserCardPopover user={user}>
      <GooseLink
        href={`/u/${user.id}`}
        className="flex min-w-0 items-center gap-1.5 text-xs hover:text-primary"
      >
        <Avatar className="size-5">
          <AvatarImage src={user.avatarUrl} alt="" />
          <AvatarFallback>
            {user.username.slice(0, 1).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <span className="text-muted-foreground">{label}</span>
        <span className="truncate font-medium">{user.username}</span>
      </GooseLink>
    </UserCardPopover>
  );
}
function PanelEmpty({
  loading = false,
  icon: Icon,
  title,
  description,
}: {
  loading?: boolean;
  icon: ComponentType;
  title: string;
  description: string;
}) {
  return (
    <Empty className="min-h-48 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          {loading ? <Spinner /> : <Icon />}
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        {description ? (
          <EmptyDescription>{description}</EmptyDescription>
        ) : null}
      </EmptyHeader>
    </Empty>
  );
}
function resolution(item: ModerationReportItem, t: Translate) {
  const value =
    item.resolution || (item.status === "rejected" ? "ignored" : "resolved");
  return t(`reports.resolutions.${value}`, { defaultValue: value });
}
function formatDate(value: string, locale: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}
function merge<T extends { id: number }>(current: T[], incoming: T[]) {
  const ids = new Set(current.map((item) => item.id));
  return [...current, ...incoming.filter((item) => !ids.has(item.id))];
}
function withId(current: Set<number>, id: number) {
  const next = new Set(current);
  next.add(id);
  return next;
}
function withoutId(current: Set<number>, id: number) {
  const next = new Set(current);
  next.delete(id);
  return next;
}
