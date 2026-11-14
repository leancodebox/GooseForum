"use client";

import { useId, useState } from "react";
import { CalendarIcon, XIcon } from "lucide-react";
import { enUS, it, ja, zhCN } from "react-day-picker/locale";
import { format, isValid, parseISO } from "date-fns";
import { cn } from "cn";
import { Button } from "./button";
import { Calendar } from "./calendar";
import { Input } from "./input";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";

interface DatePickerBaseProps {
  value: string;
  onChange(value: string): void;
  id?: string;
  locale?: string;
  placeholder: string;
  clearLabel: string;
  disabled?: boolean;
  className?: string;
}
type DatePickerProps = DatePickerBaseProps & (
  | { withTime: true; timeLabel: string }
  | { withTime?: false; timeLabel?: never }
);
const calendarLocales: Record<string, typeof enUS> = { en: enUS, zh: zhCN, ja, it };

function localDate(value: string) {
  if (!value) return undefined;
  const date = parseISO(value);
  return isValid(date) ? date : undefined;
}

export function DatePicker({ value, onChange, id, locale = "en", placeholder, clearLabel, withTime = false, timeLabel, disabled, className }: DatePickerProps) {
  const [open, setOpen] = useState(false);
  const [month, setMonth] = useState<Date>();
  const timeId = useId();
  const triggerId = id || `${timeId}-trigger`;
  const selected = localDate(value);
  const language = locale.split("-")[0];
  const calendarLocale = calendarLocales[language] || enUS;
  const formatted = selected ? new Intl.DateTimeFormat(calendarLocale.code, { dateStyle: "medium" }).format(selected) : placeholder;
  const time = selected ? format(selected, "HH:mm") : "00:00";

  function select(date: Date | undefined) {
    if (!date) return;
    onChange(`${format(date, "yyyy-MM-dd")}${withTime ? `T${time}` : ""}`);
    setOpen(false);
  }

  return <div className={cn(withTime && "flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center")}>
    <Popover open={open} onOpenChange={(next) => { if (next) setMonth(selected || new Date()); setOpen(next); }}>
      <PopoverTrigger asChild>
        <Button id={triggerId} type="button" data-slot="date-picker-trigger" variant="outline" disabled={disabled} className={cn(withTime ? "min-w-0 flex-1" : "w-full", "justify-start font-normal", !selected && "text-muted-foreground", className)}>
          <CalendarIcon className="shrink-0" /><span className="truncate">{formatted}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent data-slot="date-picker-content" aria-labelledby={triggerId} align="start" className="max-h-[min(82dvh,44rem)] w-(--radix-popover-trigger-width) max-w-[calc(100dvw-1rem)] min-w-0 overscroll-contain overflow-x-hidden overflow-y-auto p-0">
        <Calendar className="w-full" mode="single" locale={calendarLocale} selected={selected} onSelect={select} month={month} onMonthChange={setMonth} captionLayout="label" autoFocus />
      </PopoverContent>
    </Popover>
    {withTime ? <Input id={timeId} aria-label={timeLabel} type="time" step="60" value={time} disabled={!selected || disabled} className="sm:w-32" onChange={(event) => { if (selected && event.target.validity.valid && event.target.value) onChange(`${format(selected, "yyyy-MM-dd")}T${event.target.value}`); }} /> : null}
    {value ? <Button type="button" size="icon-sm" variant="ghost" aria-label={clearLabel} onClick={() => { onChange(""); setOpen(false); }}><XIcon /></Button> : null}
  </div>;
}
