import { useEffect, useState } from "react";
import type {
  DailyTraffic,
  GithubRelease,
  GooseAdminApi,
  ServerVersion,
  SiteStatistics,
} from "@gooseforum/client";
import { Badge } from "@gooseforum/react/components/ui/badge";
import { Button } from "@gooseforum/react/components/ui/button";
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@gooseforum/react/components/ui/card";
import { Input } from "@gooseforum/react/components/ui/input";
import {
  Code2,
  ExternalLink,
  FileText,
  Link,
  MessageSquare,
  Tag,
  Users,
} from "lucide-react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { DashboardTextKey } from "../dashboard-i18n";
type Text = (k: DashboardTextKey) => string;
const date = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
export default function DashboardPage({
  api,
  text,
  locale,
}: {
  api: GooseAdminApi;
  text: Text;
  locale: string;
}) {
  const [stats, setStats] = useState<SiteStatistics>();
  const [traffic, setTraffic] = useState<DailyTraffic[]>([]);
  const [version, setVersion] = useState<ServerVersion>();
  const [releases, setReleases] = useState<GithubRelease[]>([]);
  const [statsLoading, setStatsLoading] = useState(true);
  const [trafficLoading, setTrafficLoading] = useState(true);
  const [releasesLoading, setReleasesLoading] = useState(true);
  const [end, setEnd] = useState(() => date(new Date()));
  const [start, setStart] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() - 7);
    return date(d);
  });
  useEffect(() => {
    let active = true;
    void api.dashboard
      .statistics()
      .then((value) => {
        if (active) setStats(value);
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setStatsLoading(false);
      });
    void api.dashboard
      .version()
      .then((value) => {
        if (active) setVersion(value);
      })
      .catch(() => undefined);
    void api.dashboard
      .releases()
      .then((value) => {
        if (active) setReleases(value);
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setReleasesLoading(false);
      });
    return () => {
      active = false;
    };
  }, [api]);
  useEffect(() => {
    let active = true;
    setTrafficLoading(true);
    void api.dashboard
      .traffic(start, end)
      .then((value) => {
        if (active) setTraffic(value);
      })
      .catch(() => {
        if (active) setTraffic([]);
      })
      .finally(() => {
        if (active) setTrafficLoading(false);
      });
    return () => {
      active = false;
    };
  }, [api, end, start]);
  const cards = [
    {
      label: text("users"),
      value: stats?.userCount,
      delta: stats?.userMonthCount,
      icon: Users,
    },
    {
      label: text("topics"),
      value: stats?.topicMaxId,
      delta: stats?.topicMonthCount,
      icon: FileText,
    },
    { label: text("posts"), value: stats?.postMaxId, icon: MessageSquare },
    { label: text("links"), value: stats?.linksCount, icon: Link },
  ];
  return (
    <main className="flex flex-1 flex-col gap-4 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{text("title")}</h2>
          <p className="text-xs text-muted-foreground">{text("description")}</p>
        </div>
        <div className="flex items-center gap-1.5 rounded-md border bg-muted/30 px-2 py-1 text-sm">
          <span className="size-1.5 rounded-full bg-success" />
          <strong>{version?.version || "dev"}</strong>
          {version ? (
            <Badge variant="secondary">{mode(version.mode, text)}</Badge>
          ) : null}
          {version?.commit ? (
            <span className="text-xs text-muted-foreground">
              #{version.commit.slice(0, 7)}
            </span>
          ) : null}
        </div>
      </header>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map(({ label, value, delta, icon: Icon }) => (
          <Card
            key={label}
            className="min-h-32 bg-gradient-to-t from-primary/5 to-card py-4"
          >
            <CardHeader>
              <CardDescription className="flex items-center gap-2 font-medium">
                <Icon className="size-4" />
                {label}
              </CardDescription>
              <CardTitle className="text-3xl tabular-nums">
                {statsLoading || value === undefined
                  ? "…"
                  : value.toLocaleString(locale)}
              </CardTitle>
              {delta !== undefined ? (
                <CardAction>
                  <Badge variant="outline">+{delta}</Badge>
                </CardAction>
              ) : null}
            </CardHeader>
          </Card>
        ))}
      </div>
      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <section className="rounded-xl border bg-background p-4">
          <div className="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 className="font-semibold">{text("traffic")}</h3>
              <p className="text-sm text-muted-foreground">
                {text("trafficHint")}
              </p>
            </div>
            <div className="flex gap-2">
              <Input
                aria-label={text("start")}
                type="date"
                value={start}
                onChange={(e) => setStart(e.target.value)}
              />
              <Input
                aria-label={text("end")}
                type="date"
                value={end}
                onChange={(e) => setEnd(e.target.value)}
              />
            </div>
          </div>
          <div className="h-80">
            {traffic.length ? (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={traffic}>
                  <defs>
                    <linearGradient id="reg" x1="0" y1="0" x2="0" y2="1">
                      <stop
                        offset="5%"
                        stopColor="var(--chart-1)"
                        stopOpacity={0.5}
                      />
                      <stop
                        offset="95%"
                        stopColor="var(--chart-1)"
                        stopOpacity={0.05}
                      />
                    </linearGradient>
                  </defs>
                  <CartesianGrid vertical={false} stroke="var(--border)" />
                  <XAxis dataKey="date" tickLine={false} axisLine={false} />
                  <YAxis tickLine={false} axisLine={false} />
                  <Tooltip />
                  <Legend />
                  <Area
                    type="monotone"
                    dataKey="regCount"
                    name={text("registrations")}
                    stroke="var(--chart-1)"
                    fill="url(#reg)"
                  />
                  <Area
                    type="monotone"
                    dataKey="topicCount"
                    name={text("topics")}
                    stroke="var(--chart-2)"
                    fill="var(--chart-2)"
                    fillOpacity={0.12}
                  />
                  <Area
                    type="monotone"
                    dataKey="replyCount"
                    name={text("posts")}
                    stroke="var(--chart-4)"
                    fill="var(--chart-4)"
                    fillOpacity={0.1}
                  />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="grid h-full place-items-center text-muted-foreground">
                {trafficLoading ? text("loading") : text("empty")}
              </div>
            )}
          </div>
        </section>
        <section className="overflow-hidden rounded-xl border bg-background">
          <div className="flex items-start justify-between border-b p-4">
            <div>
              <h3 className="flex items-center gap-2 font-semibold">
                <Code2 className="size-4" />
                {text("releases")}
              </h3>
              <p className="text-sm text-muted-foreground">{text("version")}</p>
            </div>
            <Button asChild variant="ghost" size="sm">
              <a
                href="https://github.com/leancodebox/GooseForum/releases"
                target="_blank"
                rel="noreferrer"
              >
                {text("viewAll")}
                <ExternalLink data-icon="inline-end" />
              </a>
            </Button>
          </div>
          <div className="max-h-96 divide-y overflow-y-auto">
            {releases.length ? (
              releases.map((release) => (
                <a
                  key={release.id}
                  href={release.html_url}
                  target="_blank"
                  rel="noreferrer"
                  className="group flex items-center gap-3 p-3 hover:bg-muted/40"
                >
                  <Tag className="size-4 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <div className="font-semibold">{release.tag_name}</div>
                    <p className="truncate text-xs text-muted-foreground">
                      {release.body || "—"}
                    </p>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {new Date(release.published_at).toLocaleDateString(locale)}
                  </span>
                </a>
              ))
            ) : (
              <div className="p-8 text-center text-muted-foreground">
                {releasesLoading ? text("loading") : text("empty")}
              </div>
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
function mode(v: string, t: Text) {
  return t(
    (["release", "snapshot", "development", "custom"].includes(v)
      ? v
      : "custom") as DashboardTextKey,
  );
}
