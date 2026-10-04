import { afterEach, expect, it, vi } from 'vitest';
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { GooseSiteApi } from '@gooseforum/client';
import { GooseI18nProvider } from '@gooseforum/runtime/i18n';
import { GooseRuntimeProvider, type GooseRuntime } from '@gooseforum/runtime';
import { AuthLogSettings } from '../src/site/settings/settings-auth-logs';
import settings from '../../client/src/i18n/messages/en-settings';

afterEach(() => { cleanup(); vi.useRealTimers(); });
function setup(authLogs: unknown) {
  const view = (logs: unknown) => {
    const runtime = { api: { users: { authLogs: logs } } as unknown as GooseSiteApi, locale: 'en' } as GooseRuntime;
    return <GooseI18nProvider locale="en" initialResources={{en:{settings}}}><GooseRuntimeProvider runtime={runtime}><AuthLogSettings /></GooseRuntimeProvider></GooseI18nProvider>;
  };
  const rendered = render(view(authLogs));
  return { ...rendered, switchApi: (logs: unknown) => rendered.rerender(view(logs)) };
}
const row = (id: number) => ({ id, userId: 7, action: 'login_success', authMethod: 'password', oauthProvider: '', result: 'success', clientIp: `127.0.0.${id}`, userAgent: 'Browser', createdAt: '2026-10-03T09:00:00Z' });

it('paginates and resets the page when applying a result filter', async () => {
  const logs = vi.fn().mockImplementation(({ cursor }) => Promise.resolve({ list: [row(cursor ? 2 : 1)], nextCursor: cursor ? 0 : 101, hasMore: !cursor, pageSize: 20 }));
  setup(logs);
  await screen.findByText(/127.0.0.1/);
  const table = screen.getByRole('table');
  expect(screen.getByRole('region', { name: 'Login and security history' }).getAttribute('tabindex')).toBe('0');
  expect(within(table).getAllByRole('columnheader').map((header) => header.textContent)).toEqual(['Time', 'Event', 'Method', 'Result', 'Device']);
  expect(within(table).getByText('Success').getAttribute('data-slot')).toBe('badge');
  expect(within(table).getByText('Browser')).toBeTruthy();
  fireEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  expect(screen.getByRole('button', { name: 'Next' }).hasAttribute('disabled')).toBe(true);
  await userEvent.click(screen.getByRole('button', { name: 'Previous' }));
  await screen.findByText(/127.0.0.1/);
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  fireEvent.keyDown(screen.getByRole('combobox', { name: 'Result' }), { key: 'ArrowDown' });
  await userEvent.click(await screen.findByRole('option', { name: 'Success' }));
  await waitFor(() => expect(logs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0, result: 'success' })));
  expect(screen.getByText(/127.0.0.1/)).toBeTruthy();
});

it('ignores an older response after a filter change and exposes retryable errors', async () => {
  let finishOld: (value: unknown) => void = () => {};
  const logs = vi.fn().mockReturnValueOnce(new Promise((resolve) => { finishOld = resolve; })).mockResolvedValueOnce({ list: [row(2)], nextCursor: 0, hasMore: false }).mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce({ list: [], nextCursor: 0, hasMore: false });
  setup(logs);
  await waitFor(() => expect(logs).toHaveBeenCalledTimes(1));
  expect(screen.getByRole('status').textContent).toContain('Loading');
  fireEvent.keyDown(screen.getByRole('combobox', { name: 'Method' }), { key: 'ArrowDown' });
  await userEvent.click(await screen.findByRole('option', { name: 'OAuth' }));
  await screen.findByText(/127.0.0.2/);
  await act(async () => { finishOld({ list: [row(1)], nextCursor: 100, hasMore: true }); await Promise.resolve(); });
  expect(screen.queryByText(/127.0.0.1/)).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByRole('alert');
  expect(screen.getByRole('button', { name: 'Previous' }).hasAttribute('disabled')).toBe(true);
  expect(screen.getByRole('button', { name: 'Next' }).hasAttribute('disabled')).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText('No security events');
});

it('applies calendar dates as local day boundaries and resets pagination when cleared', async () => {
  vi.useFakeTimers({ toFake: ['Date'] });
  vi.setSystemTime(new Date(2026, 9, 4));
  const logs = vi.fn().mockImplementation(({ cursor }) => Promise.resolve({ list: [row(cursor ? 2 : 1)], nextCursor: cursor ? 0 : 101, hasMore: !cursor, pageSize: 20 }));
  setup(logs);
  await screen.findByText(/127.0.0.1/);
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!);
  expect(logs).toHaveBeenCalledTimes(2);
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/7/2026"]')!);
  await waitFor(() => expect(logs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0, since: new Date(2026, 9, 6).toISOString(), until: new Date(2026, 9, 7, 23, 59, 59, 999).toISOString() })));
  expect(logs).toHaveBeenCalledTimes(3);
  await userEvent.click(screen.getByLabelText('Date range'));
  await userEvent.click(screen.getByRole('button', { name: 'Clear date range' }));
  await waitFor(() => expect(logs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0, since: undefined, until: undefined })));
});

it('refreshes from cursor zero and clears previous pages', async () => {
  const logs = vi.fn().mockImplementation(({ cursor }) => Promise.resolve({ list: [row(cursor ? 2 : 1)], nextCursor: cursor ? 0 : 71, hasMore: !cursor }));
  setup(logs);
  await screen.findByText(/127.0.0.1/);
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText(/127.0.0.2/);
  expect(logs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 71 }));
  await userEvent.click(screen.getByRole('button', { name: 'Refresh' }));
  await screen.findByText(/127.0.0.1/);
  expect(logs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0 }));
  expect(screen.getByRole('button', { name: 'Previous' }).hasAttribute('disabled')).toBe(true);
});

it('starts a new API owner at cursor zero and ignores the old owner response', async () => {
  let finishOld!: (value: unknown) => void;
  const oldLogs = vi.fn().mockResolvedValueOnce({ list: [row(1)], nextCursor: 91, hasMore: true }).mockImplementationOnce(() => new Promise(resolve => { finishOld = resolve; }));
  const currentLogs = vi.fn().mockResolvedValue({ list: [row(3)], nextCursor: 0, hasMore: false });
  const mounted = setup(oldLogs);
  await screen.findByText(/127.0.0.1/);
  await userEvent.click(screen.getByRole('button', { name: 'Next' }));
  await waitFor(() => expect(oldLogs).toHaveBeenCalledTimes(2));
  mounted.switchApi(currentLogs);
  await screen.findByText(/127.0.0.3/);
  expect(currentLogs).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 0 }));
  await act(() => { finishOld({ list: [row(2)], nextCursor: 51, hasMore: true }); return Promise.resolve(); });
  expect(screen.queryByText(/127.0.0.2/)).toBeNull();
  expect(screen.getByRole('button', { name: 'Previous' }).hasAttribute('disabled')).toBe(true);
  expect(screen.getByRole('button', { name: 'Next' }).hasAttribute('disabled')).toBe(true);
});
