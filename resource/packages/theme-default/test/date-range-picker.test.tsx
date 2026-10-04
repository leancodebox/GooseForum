import { useState } from 'react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DateRangePicker, type DateRangeValue } from '@gooseforum/ui/components/date-range-picker';

beforeEach(() => { vi.useFakeTimers({ toFake: ['Date'] }); vi.setSystemTime(new Date(2026, 9, 4)); });
afterEach(() => { cleanup(); vi.useRealTimers(); vi.unstubAllGlobals(); });

function Harness({ initial = { from: '', to: '' }, onChange = vi.fn(), locale = 'en' }: { initial?: DateRangeValue; onChange?(value: DateRangeValue): void; locale?: string }) {
  const [value, setValue] = useState(initial);
  return <><label htmlFor="range">Date range</label><DateRangePicker id="range" value={value} locale={locale} placeholder="Choose date range" clearLabel="Clear date range" onChange={(next) => { setValue(next); onChange(next); }} /></>;
}
function day(value: string) {
  return document.querySelector<HTMLButtonElement>(`[data-day="${value}"]`)!;
}

it('commits a cross-month range only after both clicks and then clears it', async () => {
  const change = vi.fn();
  render(<Harness onChange={change} />);
  await userEvent.click(screen.getByLabelText('Date range'));
  expect(screen.queryByRole('combobox')).toBeNull();
  expect(screen.getAllByRole('grid')).toHaveLength(2);
  await userEvent.click(day('10/30/2026'));
  expect(change).not.toHaveBeenCalled();
  expect(document.querySelector('[data-slot="date-range-picker-content"]')).toBeTruthy();
  await userEvent.click(day('11/5/2026'));
  expect(change).toHaveBeenCalledExactlyOnceWith({ from: '2026-10-30', to: '2026-11-05' });
  expect(document.querySelector('[data-slot="date-range-picker-content"]')).toBeNull();
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(screen.getByRole('button', { name: 'Clear date range' }));
  expect(change).toHaveBeenLastCalledWith({ from: '', to: '' });
});

it('confirms the same day on a second click and starts a fresh range on reopening', async () => {
  const change = vi.fn();
  render(<Harness initial={{ from: '2026-10-01', to: '2026-10-04' }} onChange={change} />);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(day('10/6/2026'));
  expect(change).not.toHaveBeenCalled();
  await userEvent.click(day('10/6/2026'));
  expect(change).toHaveBeenLastCalledWith({ from: '2026-10-06', to: '2026-10-06' });
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(day('10/10/2026'));
  expect(change).toHaveBeenCalledTimes(1);
  await userEvent.click(day('10/8/2026'));
  expect(change).toHaveBeenLastCalledWith({ from: '2026-10-08', to: '2026-10-10' });
});

it('cancels an unfinished range without changing the committed dates', async () => {
  const change = vi.fn();
  render(<Harness initial={{ from: '2026-10-01', to: '2026-10-04' }} onChange={change} />);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(day('10/6/2026'));
  await userEvent.keyboard('{Escape}');
  expect(change).not.toHaveBeenCalled();
  expect(screen.getByLabelText('Date range').textContent).toContain('Oct 1, 2026 - Oct 4, 2026');
});

it('shows one calendar month on mobile and allows a cross-month range', async () => {
  vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({ matches: true, addEventListener: vi.fn(), removeEventListener: vi.fn() }));
  const change = vi.fn();
  render(<Harness onChange={change} />);
  await userEvent.click(screen.getByLabelText('Date range'));
  expect(screen.getAllByRole('grid')).toHaveLength(1);
  await userEvent.click(day('10/30/2026'));
  await userEvent.click(screen.getByRole('button', { name: 'Go to the Next Month' }));
  await userEvent.click(day('11/5/2026'));
  expect(change).toHaveBeenCalledExactlyOnceWith({ from: '2026-10-30', to: '2026-11-05' });
});

it.each([['zh', '2026/10/6', '2026/10/7'], ['ja', '2026/10/6', '2026/10/7'], ['it', '06/10/2026', '07/10/2026']])('keeps local dates in the %s locale', async (locale, from, to) => {
  const change = vi.fn();
  render(<Harness locale={locale} onChange={change} />);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(day(from));
  await userEvent.click(day(to));
  expect(change).toHaveBeenCalledExactlyOnceWith({ from: '2026-10-06', to: '2026-10-07' });
});
