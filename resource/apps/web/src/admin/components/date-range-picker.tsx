import { useId, useState } from "react";
import { format, parseISO, startOfDay } from "date-fns";
import { enUS, fr, ru, it, ja, ko, zhCN } from "date-fns/locale";
import { CalendarDays } from "lucide-react";
import type { DateRange } from "react-day-picker";
import { Button } from "@gooseforum/ui/components/button";
import { Calendar } from "@gooseforum/ui/components/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import { useIsMobile } from "@gooseforum/ui/hooks/use-mobile";
import type { DashboardTextKey } from "../dashboard-i18n";
import { presetDateRange, type DashboardDateRange } from "./date-range";

const presets = [
  { days: 7, label: "last7Days" },
  { days: 30, label: "last30Days" },
  { days: 90, label: "last90Days" },
] as const;

export function DateRangePicker({
  value,
  locale,
  text,
  onChange,
}: {
  value: DashboardDateRange;
  locale: string;
  text(key: DashboardTextKey): string;
  onChange(value: DashboardDateRange): void;
}) {
  const triggerId = useId();
  const isMobile = useIsMobile();
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<DateRange>();
  const today = startOfDay(new Date());
  const from = parseISO(value.start);
  const to = parseISO(value.end);
  const formatter = new Intl.DateTimeFormat(locale, { dateStyle: "medium" });
  const draftValue = draft?.from && draft.to
    ? { start: format(draft.from, "yyyy-MM-dd"), end: format(draft.to, "yyyy-MM-dd") }
    : undefined;
  const canApply = draftValue && (draftValue.start !== value.start || draftValue.end !== value.end);

  function changeOpen(next: boolean) {
    if (next) setDraft({ from, to });
    setOpen(next);
  }

  function commit(next: DashboardDateRange) {
    onChange(next);
    setOpen(false);
  }

  return (
    <Popover open={open} onOpenChange={changeOpen}>
      <PopoverTrigger asChild>
        <Button
          id={triggerId}
          data-slot="dashboard-date-range-trigger"
          type="button"
          variant="outline"
          className="w-full min-w-0 justify-start text-left font-normal sm:w-auto"
        >
          <CalendarDays data-icon="inline-start" />
          <span className="truncate">{formatter.formatRange(from, to)}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent
        aria-labelledby={triggerId}
        align="end"
        className="max-h-[var(--radix-popover-content-available-height)] w-auto max-w-[calc(100vw-2rem)] gap-0 overflow-y-auto p-0"
      >
        <div className="flex flex-wrap gap-1 border-b p-2">
          {presets.map(({ days, label }) => {
            const preset = presetDateRange(days, today);
            const active = draftValue?.start === preset.start && draftValue.end === preset.end;
            return (
              <Button
                key={days}
                type="button"
                variant={active ? "secondary" : "ghost"}
                aria-pressed={active}
                size="sm"
                onClick={() => commit(preset)}
              >
                {text(label)}
              </Button>
            );
          })}
        </div>
        <Calendar
          autoFocus
          mode="range"
          required
          resetOnSelect
          selected={draft}
          onSelect={setDraft}
          defaultMonth={from}
          endMonth={today}
          numberOfMonths={isMobile ? 1 : 2}
          fixedWeeks
          showOutsideDays={false}
          locale={calendarLocale(locale)}
          disabled={{ after: today }}
        />
        <div className="flex flex-wrap items-center justify-between gap-2 border-t p-2">
          <p className="text-xs text-muted-foreground" aria-live="polite">
            {draft?.from && draft.to
              ? formatter.formatRange(draft.from, draft.to)
              : draft?.from
                ? <>{formatter.format(draft.from)} · {text("selectEndDate")}</>
                : text("selectRange")}
          </p>
          <div className="ml-auto flex gap-1">
            <Button type="button" variant="ghost" size="sm" onClick={() => setOpen(false)}>
              {text("cancel")}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={!canApply}
              onClick={() => { if (canApply) commit(draftValue); }}
            >
              {text("apply")}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function calendarLocale(locale: string) {
  if (locale.startsWith("zh")) return zhCN;
  if (locale.startsWith("ja")) return ja;
  if (locale.startsWith("ru")) return ru;
  if (locale.startsWith("it")) return it;
  if (locale.startsWith("ko")) return ko;
  if (locale.startsWith("fr")) return fr;
  return enUS;
}
