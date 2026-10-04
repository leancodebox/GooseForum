"use client";

import { useId, useState } from "react";
import { CalendarIcon, XIcon } from "lucide-react";
import { enUS, it, ja, zhCN } from "react-day-picker/locale";
import { format, isValid, parseISO } from "date-fns";
import { cn } from "cn";
import { Button } from "./button";
import { Calendar } from "./calendar";
import { Input } from "./input";
import { Label } from "./label";
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
  | { withTime: true; timeLabel: string; doneLabel: string }
  | { withTime?: false; timeLabel?: never; doneLabel?: never }
);
const calendarLocales: Record<string, typeof enUS> = { en: enUS, zh: zhCN, ja, it };

function localDate(value: string) {
  if (!value) return undefined;
  const date = parseISO(value);
  return isValid(date) ? date : undefined;
}

export function DatePicker({ value, onChange, id, locale = "en", placeholder, clearLabel, withTime = false, timeLabel, doneLabel, disabled, className }: DatePickerProps) {
  const [open, setOpen] = useState(false);
  const [month, setMonth] = useState<Date>();
  const timeId = useId();
  const triggerId = id || `${timeId}-trigger`;
  const selected = localDate(value);
  const language = locale.split("-")[0];
  const calendarLocale = calendarLocales[language] || enUS;
  const formatted = selected ? new Intl.DateTimeFormat(calendarLocale.code, { dateStyle: "medium", ...(withTime ? { timeStyle: "short" as const } : {}) }).format(selected) : placeholder;
  const time = selected ? format(selected, "HH:mm") : "00:00";

  function select(date: Date | undefined) {
    if (!date) return;
    onChange(`${format(date, "yyyy-MM-dd")}${withTime ? `T${time}` : ""}`);
    if (!withTime) setOpen(false);
  }

  return <Popover open={open} onOpenChange={(next) => { if (next) setMonth(selected || new Date()); setOpen(next); }}>
    <PopoverTrigger asChild>
      <Button id={triggerId} type="button" data-slot="date-picker-trigger" variant="outline" disabled={disabled} className={cn("w-full min-w-0 justify-start font-normal", !selected && "text-muted-foreground", className)}>
        <CalendarIcon className="shrink-0" /><span className="truncate">{formatted}</span>
      </Button>
    </PopoverTrigger>
    <PopoverContent data-slot="date-picker-content" aria-labelledby={triggerId} align="start" className="max-h-[var(--radix-popover-content-available-height)] w-auto max-w-[calc(100vw-2rem)] overflow-y-auto p-0">
      <Calendar mode="single" locale={calendarLocale} selected={selected} onSelect={select} month={month} onMonthChange={setMonth} captionLayout="label" autoFocus />
      {withTime ? <div className="space-y-2 border-t px-3 py-2">
        <Label htmlFor={timeId}>{timeLabel}</Label>
        <Input id={timeId} type="time" step="60" value={time} disabled={!selected} onChange={(event) => { if (selected && event.target.validity.valid && event.target.value) onChange(`${format(selected, "yyyy-MM-dd")}T${event.target.value}`); }} />
      </div> : null}
      <div className="flex items-center justify-between gap-2 border-t p-2">
        <Button type="button" size="sm" variant="ghost" disabled={!value} onClick={() => { onChange(""); setOpen(false); }}><XIcon />{clearLabel}</Button>
        {withTime && doneLabel ? <Button type="button" size="sm" variant="outline" onClick={() => setOpen(false)}>{doneLabel}</Button> : null}
      </div>
    </PopoverContent>
  </Popover>;
}
