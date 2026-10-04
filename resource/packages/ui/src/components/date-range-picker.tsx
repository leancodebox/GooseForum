"use client";

import { useId, useState, useSyncExternalStore } from "react";
import { format, isValid, parseISO } from "date-fns";
import { CalendarIcon, XIcon } from "lucide-react";
import { type DateRange } from "react-day-picker";
import { enUS, it, ja, zhCN } from "react-day-picker/locale";
import { cn } from "cn";
import { Button } from "./button";
import { Calendar } from "./calendar";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";

export interface DateRangeValue { from: string; to: string }
interface DateRangePickerProps {
  value: DateRangeValue;
  onChange(value: DateRangeValue): void;
  id?: string;
  locale?: string;
  placeholder: string;
  clearLabel: string;
  disabled?: boolean;
  className?: string;
}
const calendarLocales: Record<string, typeof enUS> = { en: enUS, zh: zhCN, ja, it };
const mobileQuery = "(max-width: 767px)";
function subscribeToViewport(onChange: () => void) {
  const media = window.matchMedia?.(mobileQuery);
  media?.addEventListener("change", onChange);
  return () => media?.removeEventListener("change", onChange);
}
function localDate(value: string) {
  const date = parseISO(value);
  return isValid(date) ? date : undefined;
}

// Based on shadcn's date-picker-with-range example; draft prevents partial filters.
export function DateRangePicker({ value, onChange, id, locale = "en", placeholder, clearLabel, disabled, className }: DateRangePickerProps) {
  const generatedId = useId();
  const triggerId = id || `${generatedId}-trigger`;
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<DateRange>();
  const [selectingEnd, setSelectingEnd] = useState(false);
  const [month, setMonth] = useState<Date>();
  const isMobile = useSyncExternalStore(subscribeToViewport, () => window.matchMedia?.(mobileQuery).matches ?? false, () => false);
  const calendarLocale = calendarLocales[locale.split("-")[0]] || enUS;
  const selected = { from: localDate(value.from), to: localDate(value.to) };
  const dateFormatter = new Intl.DateTimeFormat(calendarLocale.code, { dateStyle: "medium" });
  const formatted = selected.from && selected.to ? `${dateFormatter.format(selected.from)} - ${dateFormatter.format(selected.to)}` : placeholder;

  function select(range: DateRange | undefined, day: Date) {
    if (!selectingEnd) {
      setDraft({ from: day });
      setSelectingEnd(true);
      return;
    }
    setDraft(range);
    if (range?.from && range.to) {
      onChange({ from: format(range.from, "yyyy-MM-dd"), to: format(range.to, "yyyy-MM-dd") });
      setOpen(false);
    }
  }

  return <Popover open={open} onOpenChange={(next) => {
    if (next) { setDraft(selected.from ? selected : undefined); setSelectingEnd(false); setMonth(selected.from || new Date()); }
    setOpen(next);
  }}>
    <PopoverTrigger asChild>
      <Button id={triggerId} type="button" data-slot="date-range-picker-trigger" variant="outline" disabled={disabled} className={cn("w-full min-w-0 justify-start text-left font-normal", !selected.from && "text-muted-foreground", className)}>
        <CalendarIcon /><span className="truncate">{formatted}</span>
      </Button>
    </PopoverTrigger>
    <PopoverContent data-slot="date-range-picker-content" aria-labelledby={triggerId} align="start" className="max-h-[var(--radix-popover-content-available-height)] w-auto max-w-[calc(100vw-2rem)] overflow-y-auto p-0">
      <Calendar autoFocus mode="range" min={0} resetOnSelect defaultMonth={selected.from} month={month} onMonthChange={setMonth} selected={draft} onSelect={select} numberOfMonths={isMobile ? 1 : 2} locale={calendarLocale} captionLayout="label" />
      <div className="border-t p-2"><Button type="button" size="sm" variant="ghost" disabled={!value.from && !value.to} onClick={() => { onChange({ from: "", to: "" }); setOpen(false); }}><XIcon />{clearLabel}</Button></div>
    </PopoverContent>
  </Popover>;
}
