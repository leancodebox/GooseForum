import { useMemo, type CSSProperties } from "react";
import { parseISO } from "date-fns";
import type { DailyTraffic } from "@gooseforum/client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@gooseforum/ui/components/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@gooseforum/ui/components/chart";
import { Skeleton } from "@gooseforum/ui/components/skeleton";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import type { DashboardTextKey } from "../dashboard-i18n";
import { type DashboardDateRange } from "./date-range";
import { DateRangePicker } from "./date-range-picker";

type Text = (key: DashboardTextKey) => string;

export function TrafficOverview({
  data,
  loading,
  range,
  locale,
  text,
  onRangeChange,
}: {
  data: DailyTraffic[];
  loading: boolean;
  range: DashboardDateRange;
  locale: string;
  text: Text;
  onRangeChange(value: DashboardDateRange): void;
}) {
  const config = useMemo(
    () =>
      ({
        regCount: { label: text("registrations"), color: "var(--chart-1)" },
        topicCount: { label: text("topics"), color: "var(--chart-2)" },
        replyCount: { label: text("posts"), color: "var(--chart-4)" },
      }) satisfies ChartConfig,
    [text],
  );
  const totals = useMemo(
    () =>
      data.reduce(
        (result, item) => ({
          regCount: result.regCount + item.regCount,
          topicCount: result.topicCount + item.topicCount,
          replyCount: result.replyCount + item.replyCount,
        }),
        { regCount: 0, topicCount: 0, replyCount: 0 },
      ),
    [data],
  );
  const dateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { month: "short", day: "numeric" }),
    [locale],
  );
  const longDateFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat(locale, {
        year: "numeric",
        month: "short",
        day: "numeric",
      }),
    [locale],
  );

  return (
    <Card className="h-full gap-0 overflow-hidden py-0">
      <CardHeader className="border-b px-4 py-4 sm:flex sm:flex-row sm:items-center sm:justify-between sm:px-5">
        <div className="flex flex-col gap-1">
          <CardTitle>{text("traffic")}</CardTitle>
          <CardDescription>{text("trafficHint")}</CardDescription>
        </div>
        <DateRangePicker
          value={range}
          locale={locale}
          text={text}
          onChange={onRangeChange}
        />
      </CardHeader>
      <div className="grid grid-cols-3 divide-x border-b bg-muted/20">
        {(
          [
            ["regCount", text("registrations")],
            ["topicCount", text("topics")],
            ["replyCount", text("posts")],
          ] as const
        ).map(([key, label]) => (
          <div
            key={key}
            className="flex min-w-0 flex-col gap-1 px-3 py-3 sm:px-5"
          >
            <div className="flex items-center gap-2 text-xs text-muted-foreground sm:text-sm">
              <span
                className="size-2 shrink-0 rounded-full bg-(--metric-color)"
                style={{ "--metric-color": config[key].color } as CSSProperties}
              />
              <span className="truncate">{label}</span>
            </div>
            {loading ? (
              <Skeleton className="h-7 w-16" />
            ) : (
              <strong className="text-xl tabular-nums tracking-tight sm:text-2xl">
                {totals[key].toLocaleString(locale)}
              </strong>
            )}
          </div>
        ))}
      </div>
      <CardContent className="px-2 pb-4 pt-5 sm:px-4">
        {loading ? (
          <div className="flex h-[290px] flex-col gap-4 px-4 py-3 sm:h-[330px]">
            <Skeleton className="h-full w-full" />
          </div>
        ) : data.length ? (
          <ChartContainer
            config={config}
            className="h-[290px] w-full sm:h-[330px]"
          >
            <AreaChart
              accessibilityLayer
              data={data}
              margin={{ top: 8, right: 12, left: 4, bottom: 0 }}
            >
              <defs>
                {Object.entries(config).map(([key, item]) => (
                  <linearGradient
                    key={key}
                    id={`traffic-${key}`}
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop
                      offset="8%"
                      stopColor={item.color}
                      stopOpacity={0.28}
                    />
                    <stop
                      offset="92%"
                      stopColor={item.color}
                      stopOpacity={0.02}
                    />
                  </linearGradient>
                ))}
              </defs>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={10}
                minTickGap={28}
                tickFormatter={(value) =>
                  dateFormatter.format(parseISO(String(value)))
                }
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                allowDecimals={false}
                width={36}
              />
              <ChartTooltip
                cursor={{ stroke: "var(--border)", strokeDasharray: "4 4" }}
                content={
                  <ChartTooltipContent
                    indicator="line"
                    labelFormatter={(value) =>
                      longDateFormatter.format(parseISO(String(value)))
                    }
                  />
                }
              />
              <Area
                type="monotone"
                dataKey="replyCount"
                stroke="var(--color-replyCount)"
                fill="url(#traffic-replyCount)"
                strokeWidth={2}
              />
              <Area
                type="monotone"
                dataKey="topicCount"
                stroke="var(--color-topicCount)"
                fill="url(#traffic-topicCount)"
                strokeWidth={2}
              />
              <Area
                type="monotone"
                dataKey="regCount"
                stroke="var(--color-regCount)"
                fill="url(#traffic-regCount)"
                strokeWidth={2}
              />
            </AreaChart>
          </ChartContainer>
        ) : (
          <div className="grid h-[290px] place-items-center text-sm text-muted-foreground sm:h-[330px]">
            {text("empty")}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
