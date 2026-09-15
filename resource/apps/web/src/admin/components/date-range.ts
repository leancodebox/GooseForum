import { format, subDays } from "date-fns";

export interface DashboardDateRange {
  start: string;
  end: string;
}

export function presetDateRange(days: number, today: Date): DashboardDateRange {
  const end = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  return {
    start: format(subDays(end, days - 1), "yyyy-MM-dd"),
    end: format(end, "yyyy-MM-dd"),
  };
}
