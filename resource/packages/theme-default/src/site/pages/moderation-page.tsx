import {
  useCallback,
  useEffect,
  useState,
  type ComponentType,
  type ReactNode,
} from "react";
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
import { Alert, AlertDescription } from "@gooseforum/ui/components/alert";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { Button } from "@gooseforum/ui/components/button";
import { ListFilter, ListFilterItem } from "@gooseforum/ui/components/list-filter";
import { Badge } from "@gooseforum/ui/components/badge";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { Spinner } from "@gooseforum/ui/components/spinner";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@gooseforum/ui/components/tabs";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { PageHeader } from "../layout/page-header";
import { UserCardPopover } from "../users/user-card-popover";

type ConsoleTab = "reports" | "ban" | "logs" | "guidance";
type ReportStatus = "open" | "closed";
const consoleTabs: Array<{ key: ConsoleTab; icon: ComponentType<{ className?: string }> }> = [
  { key: "reports", icon: Flag },
  { key: "ban", icon: Ban },
  { key: "logs", icon: History },
  { key: "guidance", icon: Scale },
];

export function ModerationPageView({ page }: { page: ModerationPageProps }) {
  const { t } = useTranslation("moderation");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
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
          serverError(reason, t("reports.loadFailed")),
        );
      } finally {
        setReportLoading(false);
      }
    },
    [reportCursor, reportLoading, reportStatus, runtime.api.moderation, serverError, t],
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
          serverError(reason, t("logs.loadFailed")),
        );
      } finally {
        setLogLoading(false);
      }
    },
    [logCursor, logLoading, runtime.api.moderation, serverError, t],
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
        serverError(reason, t("reports.actionFailed")),
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
        serverError(reason, t("blocked.loadFailed")),
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
        className="gap-0 px-4 lg:px-0"
      >
        <div className="min-w-0">
          <div
            data-slot="moderation-tabs-frame"
            className="border-b"
          >
            <TabsList
              variant="line"
              aria-label={t("title")}
              className="h-auto min-h-9 w-full max-w-full justify-start gap-1 rounded-none group-data-horizontal/tabs:h-auto sm:gap-4"
            >
              {consoleTabs.map(({ key, icon: Icon }) => (
                <TabsTrigger
                  key={key}
                  value={key}
                  className="h-auto min-h-8 min-w-0 flex-1 gap-1.5 whitespace-normal px-1 text-center text-xs leading-6 sm:px-3 sm:text-sm"
                >
                  <Icon className="hidden size-4 shrink-0 sm:block" />
                  {t(`tabs.${key}`)}
                </TabsTrigger>
              ))}
            </TabsList>
          </div>
          <TabsContent value="reports" className="mt-0">
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
              onRefresh={() => void loadReports(true)}
              onAction={(item, action) => void handleReport(item, action)}
            />
          </TabsContent>
          <TabsContent value="ban" className="mt-0 pt-5">
            <BlockedPanel
              page={page}
              topics={topics}
              busy={topicBusy}
              error={topicError}
              t={t}
              onRestore={(topic) => void restoreTopic(topic)}
            />
          </TabsContent>
          <TabsContent value="logs" className="mt-0 pt-1">
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
          <TabsContent value="guidance" className="mt-0 pt-5">
            <Guidance t={t} />
          </TabsContent>
        </div>
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
  onRefresh,
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
  onRefresh(): void;
  onAction(item: ModerationReportItem, action: "ban" | "reject"): void;
}) {
  return (
    <section className="flex flex-col">
      {error ? (
        <Alert variant="destructive" className="m-3 mb-0">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="min-w-0">
        <div className="flex items-center justify-between gap-4 py-5">
          <Tabs value={status} onValueChange={(value) => onStatus(value as ReportStatus)} className="gap-0">
            <ListFilter asChild>
              <TabsList aria-label={t("tabs.reports")}>
                <ListFilterItem asChild>
                  <TabsTrigger value="open">{t("reports.statusTabs.open")}</TabsTrigger>
                </ListFilterItem>
                <ListFilterItem asChild>
                  <TabsTrigger value="closed">{t("reports.statusTabs.closed")}</TabsTrigger>
                </ListFilterItem>
              </TabsList>
            </ListFilter>
          </Tabs>
          <Button
            variant="outline"
            size="icon-sm"
            aria-label={t("reports.refresh")}
            title={t("reports.refresh")}
            disabled={loading}
            onClick={onRefresh}
          >
            {loading ? <Spinner /> : <RotateCcw />}
          </Button>
        </div>
        {items.length ? (
          <div className="divide-y border-t" aria-busy={loading}>
            {items.map((item) => (
              <article
                key={item.id}
                className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-3 gap-y-2 py-4 hover:bg-muted/40 sm:px-3 lg:grid-cols-[minmax(0,1fr)_140px_130px_150px] lg:items-center"
                aria-busy={busy.has(item.id)}
              >
                <div className="col-span-2 flex min-w-0 items-start gap-3 lg:col-span-1">
                <span className="flex size-7 shrink-0 items-center justify-center rounded bg-muted text-muted-foreground">
                  <Flag className="size-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex min-h-7 min-w-0 flex-wrap items-center gap-2">
                    <span className="text-xs text-muted-foreground">
                      {t(`reports.targetTypes.${item.targetType}`)}
                    </span>
                    <Badge variant="secondary" className="text-xs">
                      {t(`reports.reasons.${item.reason}`, { defaultValue: item.reason })}
                    </Badge>
                    {item.categories.map((category) => (
                      <GooseLink
                        href={category.url}
                        key={category.id}
                        className="inline-flex items-center gap-1 text-xs text-muted-foreground"
                      >
                        <span
                          className="size-2 rounded-sm"
                          style={{ backgroundColor: category.color }}
                        />
                        {category.name}
                      </GooseLink>
                    ))}
                  </div>
                  <GooseLink href={item.targetUrl} className="block text-sm font-medium leading-6 hover:text-primary">
                    {item.title}
                  </GooseLink>
                  <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                    {item.excerpt || t("reports.noExcerpt")}
                  </p>
                  {item.note && (
                    <p className="mt-1 text-xs leading-5">
                      <span className="text-muted-foreground">{t("reports.note")}: </span>{item.note}
                    </p>
                  )}
                  {status === "closed" && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      {t("reports.submitted")} {formatDate(item.createdAt, locale)}
                    </p>
                  )}
                </div>
                </div>
                <div className="grid min-w-0 gap-2 pl-10 lg:pl-0">
                  <Person user={item.reporter} label={t("reports.reporter")} />
                  {status === "closed" && item.handler.id ? (
                    <Person user={item.handler} label={t("reports.handler")} />
                  ) : null}
                </div>
                <time dateTime={status === "closed" ? item.handledAt || item.createdAt : item.createdAt} className="col-start-1 pl-10 text-xs text-muted-foreground lg:col-start-auto lg:pl-0">{formatDate(status === "closed" ? item.handledAt || item.createdAt : item.createdAt, locale)}</time>
                <div className="col-start-2 row-start-2 row-end-4 flex flex-wrap items-center justify-end gap-2 lg:col-start-auto lg:row-auto">
                  {status === "open" ? (
                    <>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={busy.has(item.id)}
                        onClick={() => onAction(item, "ban")}
                      >
                        {busy.has(item.id) ? <Spinner /> : <Ban data-icon="inline-start" />}
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
                    <span className="text-xs text-muted-foreground">{resolution(item, t)}</span>
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
          <PanelFooter>
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
          </PanelFooter>
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
    <section className="flex flex-col">
      <div className="mb-4 flex flex-wrap gap-1">
        {page.categoryTabs.map((tab) => (
          <Button
            key={tab.key}
            asChild
            size="sm"
            variant={tab.active ? "outline" : "ghost"}
            className={
              tab.active
                ? "font-semibold text-foreground shadow-none"
                : "text-muted-foreground hover:text-foreground"
            }
            style={tab.active ? {
              backgroundColor: `color-mix(in srgb, ${tab.color || "var(--primary)"} 14%, var(--background))`,
              borderColor: tab.color || "var(--primary)",
            } : undefined}
          >
            <GooseLink
              href={tab.url}
              aria-current={tab.active ? "page" : undefined}
            >
              <span aria-hidden="true" className="size-2 shrink-0 rounded-sm" style={{ backgroundColor: tab.color || "var(--primary)" }} />
              {tab.label}
            </GooseLink>
          </Button>
        ))}
      </div>
      {error ? (
        <Alert variant="destructive" className="m-3 mb-0">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="overflow-hidden">
        {topics.length ? (
          <div className="divide-y">
            {topics.map((topic) => (
              <article
                key={topic.id}
                className="flex items-center justify-between gap-4 py-4 hover:bg-muted/40 sm:px-3"
              >
                <div className="min-w-0">
                  <GooseLink
                    href={topic.url}
                    className="block text-sm font-medium leading-6 hover:text-primary"
                  >
                    {topic.title}
                  </GooseLink>
                  <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                    {topic.description}
                  </p>
                  <div className="mt-1 flex flex-wrap gap-2">{topic.categories.map((category) => <GooseLink key={category.id} href={category.url} className="inline-flex items-center gap-1 text-xs text-muted-foreground"><span className="size-2 rounded-sm" style={{ backgroundColor: category.color }} />{category.name}</GooseLink>)}</div>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  className="shrink-0"
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
          <PanelFooter>
            <Button asChild variant="outline" size="sm">
              <GooseLink href={page.pagination.nextUrl}>
                {t("blocked.next")}
              </GooseLink>
            </Button>
          </PanelFooter>
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
    <section className="flex flex-col">
      {error ? (
        <Alert variant="destructive" className="m-3 mb-0">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="overflow-hidden">
        {items.length ? (
          <div className="divide-y">
            {items.map((item) => (
              <article
                key={item.id}
                className="grid grid-cols-[28px_minmax(0,1fr)] gap-x-3 gap-y-2 py-3.5 hover:bg-muted/40 sm:px-3 lg:grid-cols-[28px_minmax(0,1fr)_130px]"
              >
                <span className="flex h-7 items-center justify-center text-muted-foreground">
                  <History className="size-4" />
                </span>
                <div className="min-w-0">
                  <div className="flex min-h-7 flex-wrap items-center gap-1.5 text-sm leading-7">
                    <Person user={item.actor} inline />
                    <span className="text-muted-foreground">
                      {t(`logs.actions.${item.action}`, {
                        defaultValue: t("logs.actions.operation"),
                      })}
                    </span>{" "}
                    {item.subject.url ? (
                      <GooseLink
                        href={item.subject.url}
                        className="font-medium hover:text-primary"
                      >
                        {item.subject.title}
                      </GooseLink>
                    ) : (
                      <strong className="font-medium">{item.subject.title}</strong>
                    )}
                  </div>
                  {item.subject.excerpt ? (
                    <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                      {item.subject.excerpt}
                    </p>
                  ) : null}
                  <div className="mt-1 flex flex-wrap gap-2">{item.categories.map((category) => <GooseLink key={category.id} href={category.url} className="inline-flex items-center gap-1 text-xs text-muted-foreground"><span className="size-2 rounded-sm" style={{ backgroundColor: category.color }} />{category.name}</GooseLink>)}</div>
                </div>
                <time dateTime={item.createdAt} className="col-start-2 text-xs leading-7 text-muted-foreground lg:col-start-auto lg:text-right">
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
          <PanelFooter>
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
          </PanelFooter>
        ) : null}
      </div>
    </section>
  );
}
function PanelFooter({ children }: { children: ReactNode }) {
  return <footer className="flex min-h-12 flex-wrap items-center gap-3 pt-4 text-xs text-muted-foreground">{children}</footer>;
}

function Guidance({ t }: { t: Translate }) {
  return (
    <section>
      <Alert className="border-warning/30 bg-warning/10">
        <CircleAlert className="h-6 w-4 self-start text-warning translate-y-0 row-span-1" />
        <AlertDescription className="leading-6 text-foreground/75">
          {t("notice")}
        </AlertDescription>
      </Alert>
      <div className="mt-3 divide-y">
        {(["rule", "context", "restraint"] as const).map((key) => (
          <article key={key} className="py-4">
            <h3 className="text-sm font-semibold">
              {t(`guidance.${key}.title`)}
            </h3>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">
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
  inline = false,
}: {
  user: ModerationReportItem["reporter"];
  label?: string;
  inline?: boolean;
}) {
  return (
    <UserCardPopover user={user}>
      <GooseLink
        href={`/u/${user.id}`}
        className={`flex min-w-0 items-center gap-1.5 hover:text-primary ${inline ? "text-sm leading-7" : "text-xs"}`}
      >
        <Avatar className="size-5 shrink-0">
          <AvatarImage src={user.avatarUrl} alt="" />
          <AvatarFallback>
            {user.username.slice(0, 1).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        {label && <span className="shrink-0 text-muted-foreground">{label}</span>}
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
