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
            className="flex min-w-0 items-center gap-1.5 px-2 py-2 sm:gap-2 sm:px-4 sm:py-2.5"
          >
            <span
              className="size-1.5 shrink-0 rounded-full bg-(--metric-color) sm:size-2"
              style={{ "--metric-color": config[key].color } as CSSProperties}
            />
            <span className="truncate text-xs text-muted-foreground sm:text-sm">{label}</span>
            {loading ? (
              <Skeleton className="h-6 w-4 shrink-0 sm:h-7 sm:w-6" />
            ) : (
              <strong className="shrink-0 text-base tabular-nums tracking-tight sm:text-xl">
                {totals[key].toLocaleString(locale)}
              </strong>
            )}
          </div>
        ))}
      </div>
      <CardContent className="h-[306px] px-2 pb-4 pt-5 sm:h-[336px] sm:px-4">
        {loading ? (
          <div className="h-full px-4 py-3">
            <Skeleton className="h-full w-full" />
          </div>
        ) : data.length ? (
          <ChartContainer
            config={config}
            className="h-full w-full"
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
          <div className="grid h-full place-items-center text-sm text-muted-foreground">
            {text("empty")}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
