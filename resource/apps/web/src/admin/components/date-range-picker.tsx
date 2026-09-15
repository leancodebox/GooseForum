import { useState } from "react";
import { format, parseISO, subMonths } from "date-fns";
import { enUS, it, ja, zhCN } from "date-fns/locale";
import { CalendarDays } from "lucide-react";
import type { DateRange } from "react-day-picker";
import { Button } from "@gooseforum/ui/components/button";
import { Calendar } from "@gooseforum/ui/components/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import type { DashboardTextKey } from "../dashboard-i18n";
import { presetDateRange, type DashboardDateRange } from "./date-range";

type Text = (key: DashboardTextKey) => string;

export function DateRangePicker({
  value,
  locale,
  text,
  onChange,
}: {
  value: DashboardDateRange;
  locale: string;
  text: Text;
  onChange(value: DashboardDateRange): void;
}) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<DateRange>(() => toDateRange(value));
  const today = startOfLocalDay(new Date());

  function changeOpen(next: boolean) {
    if (next) setDraft(toDateRange(value));
    setOpen(next);
  }

  function apply() {
    if (!draft.from || !draft.to) return;
    onChange({
      start: format(draft.from, "yyyy-MM-dd"),
      end: format(draft.to, "yyyy-MM-dd"),
    });
    setOpen(false);
  }

  function applyPreset(days: number) {
    const next = presetDateRange(days, today);
    setDraft(toDateRange(next));
    onChange(next);
    setOpen(false);
  }

  return (
    <Popover open={open} onOpenChange={changeOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className="w-full justify-start font-normal sm:w-auto"
        >
          <CalendarDays data-icon="inline-start" />
          <span>{rangeLabel(value, locale)}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        className="max-h-[min(82vh,44rem)] w-auto max-w-[calc(100vw-2rem)] overflow-y-auto p-0"
      >
        <div className="flex flex-wrap gap-2 border-b p-3">
          {[
            [7, text("last7Days")],
            [30, text("last30Days")],
            [90, text("last90Days")],
          ].map(([days, label]) => (
            <Button
              key={days}
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => applyPreset(Number(days))}
            >
              {label}
            </Button>
          ))}
        </div>
        <Calendar
          mode="range"
          selected={draft}
          onSelect={(range) =>
            setDraft(range || { from: undefined, to: undefined })
          }
          defaultMonth={subMonths(draft.to || today, 1)}
          numberOfMonths={2}
          locale={calendarLocale(locale)}
          disabled={{ after: today }}
        />
        <div className="flex items-center justify-between gap-3 border-t p-3">
          <p className="text-xs text-muted-foreground">
            {draft.from && draft.to
              ? rangeLabel(
                  {
                    start: format(draft.from, "yyyy-MM-dd"),
                    end: format(draft.to, "yyyy-MM-dd"),
                  },
                  locale,
                )
              : text("selectRange")}
          </p>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setOpen(false)}
            >
              {text("cancel")}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={!draft.from || !draft.to}
              onClick={apply}
            >
              {text("apply")}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function toDateRange(value: DashboardDateRange): DateRange {
  return { from: parseISO(value.start), to: parseISO(value.end) };
}

function rangeLabel(value: DashboardDateRange, locale: string) {
  const formatter = new Intl.DateTimeFormat(locale, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
  return `${formatter.format(parseISO(value.start))} – ${formatter.format(parseISO(value.end))}`;
}

function calendarLocale(locale: string) {
  if (locale.startsWith("zh")) return zhCN;
  if (locale.startsWith("ja")) return ja;
  if (locale.startsWith("it")) return it;
  return enUS;
}

function startOfLocalDay(value: Date) {
  return new Date(value.getFullYear(), value.getMonth(), value.getDate());
}
