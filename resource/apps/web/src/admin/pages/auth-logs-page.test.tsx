import { act, cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, it, vi } from 'vitest';
import type { GooseAdminApi } from '@gooseforum/client';
import { AuthLogsPage } from './auth-logs-page';
import { TooltipProvider } from '@gooseforum/ui/components/tooltip';

afterEach(() => { cleanup(); vi.useRealTimers(); });
vi.stubGlobal('ResizeObserver', class { observe() {} unobserve() {} disconnect() {} });

it('applies a complete local date range atomically and clears both boundaries', async () => {
  vi.useFakeTimers({ toFake: ['Date'] });
  vi.setSystemTime(new Date(2026, 9, 4));
  const authLogs = vi.fn().mockResolvedValue({ list: [], nextCursor: 77, hasMore: true, pageSize: 20 });
  const api = { audit: { authLogs } } as unknown as GooseAdminApi;
  render(<TooltipProvider><AuthLogsPage api={api} locale="en" /></TooltipProvider>);
  await screen.findByText('No security events');
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await waitFor(() => expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 77 })));
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!);
  expect(authLogs).toHaveBeenCalledTimes(2);
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/7/2026"]')!);
  await waitFor(() => expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0, since: new Date(2026, 9, 6).toISOString(), until: new Date(2026, 9, 7, 23, 59, 59, 999).toISOString() })));
  expect(authLogs).toHaveBeenCalledTimes(3);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(screen.getByRole('button', { name: 'Clear date range' }));
  await waitFor(() => expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ since: undefined, until: undefined })));
});

const row = (id: number) => ({ id, userId: 7, action: 'login_success', authMethod: 'password', result: 'success', clientIp: `127.0.0.${id}`, createdAt: '2026-10-03T09:00:00Z' });

it('uses server cursors for next and previous, disables the last page, and resets on refresh or user filter', async () => {
  const authLogs = vi.fn().mockImplementation(({ cursor }) => Promise.resolve({ list: [row(cursor ? 2 : 1)], nextCursor: cursor ? 0 : 150, hasMore: !cursor }));
  const api = { audit: { authLogs } } as unknown as GooseAdminApi;
  render(<TooltipProvider><AuthLogsPage api={api} locale="en" /></TooltipProvider>);
  await screen.findByText(/127.0.0.1/);
  const table = screen.getByRole('table');
  expect(screen.getByRole('region', { name: 'Login and security history' }).getAttribute('tabindex')).toBe('0');
  expect(within(table).getAllByRole('columnheader').map((header) => header.textContent)).toEqual(['Time', 'Event', 'User ID', 'Method', 'Result', 'Device']);
  expect(within(table).getByText('Success').getAttribute('data-slot')).toBe('badge');
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 150, pageSize: 20 }));
  expect(screen.getByRole('button', { name: 'Next' }).hasAttribute('disabled')).toBe(true);
  await userEvent.click(screen.getByRole('button', { name: 'Previous' }));
  await screen.findByText(/127.0.0.1/);
  expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0 }));
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText(/127.0.0.1/);
  expect(screen.getByRole('button', { name: 'Previous' }).hasAttribute('disabled')).toBe(true);
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  await userEvent.type(screen.getByLabelText('User ID'), '7');
  await waitFor(() => expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0, userId: 7 })));
});

it('ignores an outdated cursor response, blocks paging after errors, and retries from the first page', async () => {
  let finishOld!: (value: unknown) => void;
  const authLogs = vi.fn().mockResolvedValueOnce({ list: [row(1)], nextCursor: 150, hasMore: true }).mockImplementationOnce(() => new Promise(resolve => { finishOld = resolve; })).mockResolvedValueOnce({ list: [row(3)], nextCursor: 55, hasMore: true }).mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce({ list: [], nextCursor: 0, hasMore: false });
  const api = { audit: { authLogs } } as unknown as GooseAdminApi;
  render(<TooltipProvider><AuthLogsPage api={api} locale="en" /></TooltipProvider>);
  await screen.findByText(/127.0.0.1/);
  const originalTable = screen.getByRole('table');
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await waitFor(() => expect(authLogs).toHaveBeenCalledTimes(2));
  expect(screen.getByRole('status').textContent).toContain('Loading');
  expect(screen.getByRole('table')).toBe(originalTable);
  expect(screen.getByText(/127.0.0.1/)).toBeTruthy();
  expect(screen.queryByText('21-21')).toBeNull();
  await userEvent.type(screen.getByLabelText('User ID'), '7');
  await screen.findByText(/127.0.0.3/);
  await act(() => { finishOld({ list: [row(2)], nextCursor: 100, hasMore: true }); return Promise.resolve(); });
  expect(screen.queryByText(/127.0.0.2/)).toBeNull();
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByRole('alert');
  expect(screen.getByRole('button', { name: 'Previous' }).hasAttribute('disabled')).toBe(true);
  expect(screen.getByRole('button', { name: 'Next' }).hasAttribute('disabled')).toBe(true);
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText('No security events');
  expect(authLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0 }));
});
