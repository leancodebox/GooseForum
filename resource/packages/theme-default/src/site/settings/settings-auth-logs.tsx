import { useEffect, useRef, useState } from 'react';
import type { AuthLogPayload } from '@gooseforum/client';
import { useGooseRuntime } from '@gooseforum/runtime';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight, History, RefreshCw } from 'lucide-react';
import { Button } from '@gooseforum/ui/components/button';
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert';
import { Badge } from '@gooseforum/ui/components/badge';
import { Empty, EmptyHeader, EmptyTitle } from '@gooseforum/ui/components/empty';
import { Spinner } from '@gooseforum/ui/components/spinner';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table';
import { DateRangePicker } from '@gooseforum/ui/components/date-range-picker';
import { Field, FieldLabel } from '@gooseforum/ui/components/field';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select';
import { SettingsSectionHeader } from './settings-section-header';

export function AuthLogSettings() {
  const runtime = useGooseRuntime();
  const { t } = useTranslation('settings');
  const [rows, setRows] = useState<AuthLogPayload[]>([]);
  const [cursors, setCursors] = useState([0]);
  const [nextCursor, setNextCursor] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [result, setResult] = useState('');
  const [method, setMethod] = useState('');
  const [range, setRange] = useState({ from: '', to: '' });
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const requestEpoch = useRef(0);
  const [owner, setOwner] = useState(runtime.api.users);
  if (owner !== runtime.api.users) {
    setOwner(runtime.api.users); setCursors([0]); setRows([]); setNextCursor(0); setHasMore(false);
    setResult(''); setMethod(''); setRange({ from: '', to: '' });
    setLoading(true); setError(false);
  }
  const cursor = cursors[cursors.length - 1];
  function reset() { requestEpoch.current++; setCursors([0]); setRefresh((value) => value + 1); setLoading(true); setError(false); }
  useEffect(() => {
    let active = true;
    const epoch = ++requestEpoch.current;
    const isCurrent = () => active && requestEpoch.current === epoch;
    void Promise.resolve().then(() => {
      if (!isCurrent()) return;
      setLoading(true); setError(false);
      return runtime.api.users.authLogs({ cursor, pageSize: 20, result, method,
      since: range.from ? new Date(`${range.from}T00:00:00`).toISOString() : undefined,
      until: range.to ? new Date(`${range.to}T23:59:59.999`).toISOString() : undefined,
      });
    }).then((data) => { if (isCurrent() && data) { setRows(data.list); setNextCursor(data.nextCursor); setHasMore(data.hasMore); } })
      .catch(() => { if (isCurrent()) setError(true); })
      .finally(() => { if (isCurrent()) setLoading(false); });
    return () => { active = false; };
  }, [runtime.api.users, cursor, result, method, range, refresh]);
  const label = (key: string) => t(`authLogs.${key}`, { defaultValue: key });
  return <section className="min-w-0">
    <SettingsSectionHeader icon={History} title={label('title')} actions={<Button type="button" variant="outline" size="icon" title={label('refresh')} aria-label={label('refresh')} disabled={loading} onClick={reset}><RefreshCw className="size-4" /></Button>} />
    <div className="min-w-0 space-y-3 p-4">
      <div className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,14rem),1fr))] items-end gap-3">
        <Field className="min-w-0"><FieldLabel htmlFor="security-log-result">{label('result')}</FieldLabel><Select value={result || 'all'} onValueChange={(value) => { setResult(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="security-log-result" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'success', 'failure', 'challenge'].map((value) => <SelectItem key={value} value={value}>{label(value)}</SelectItem>)}</SelectContent></Select></Field>
        <Field className="min-w-0"><FieldLabel htmlFor="security-log-method">{label('method')}</FieldLabel><Select value={method || 'all'} onValueChange={(value) => { setMethod(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="security-log-method" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'password', 'oauth', 'mfa', 'operator'].map((value) => <SelectItem key={value} value={value}>{label(value)}</SelectItem>)}</SelectContent></Select></Field>
        <Field className="min-w-0"><FieldLabel htmlFor="security-log-range">{label('dateRange')}</FieldLabel><DateRangePicker id="security-log-range" locale={runtime.locale} value={range} placeholder={label('rangePlaceholder')} clearLabel={label('clearRange')} onChange={(value) => { setRange(value); reset(); }} /></Field>
      </div>
      {error ? <Alert><AlertDescription>{label('error')}</AlertDescription></Alert> : loading ? <div role="status" className="flex items-center gap-2 py-4 text-sm text-muted-foreground"><Spinner role="presentation" aria-hidden="true" />{label('loading')}</div> : rows.length === 0 ? <Empty className="min-h-20 rounded-none p-4"><EmptyHeader><EmptyTitle className="font-normal tracking-normal text-muted-foreground">{label('empty')}</EmptyTitle></EmptyHeader></Empty> : <div className="min-w-0 max-w-full border-y">
        <Table className="min-w-[42rem] table-fixed" containerProps={{ tabIndex: 0, role: 'region', 'aria-label': label('title'), className: 'focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2' }}>
          <TableHeader><TableRow><TableHead className="w-40">{label('time')}</TableHead><TableHead className="w-56">{label('event')}</TableHead><TableHead className="w-36">{label('method')}</TableHead><TableHead className="w-24">{label('result')}</TableHead><TableHead className="w-56">{label('device')}</TableHead></TableRow></TableHeader>
          <TableBody>{rows.map((row) => <TableRow key={row.id}>
            <TableCell className="whitespace-normal text-xs"><time dateTime={row.createdAt}>{new Date(row.createdAt).toLocaleString(runtime.locale)}</time></TableCell>
            <TableCell className="whitespace-normal break-words font-medium">{t(`authLogActions.${row.action}`, { defaultValue: row.action })}</TableCell>
            <TableCell className="whitespace-normal break-words">{row.oauthProvider || label(row.authMethod || 'all')}</TableCell>
            <TableCell><Badge variant={row.result === 'success' ? 'secondary' : 'outline'}>{label(row.result || 'success')}</Badge></TableCell>
            <TableCell className="whitespace-normal break-all text-xs"><p>{row.clientIp || '-'}</p>{row.userAgent && <p className="mt-1 text-muted-foreground">{row.userAgent}</p>}</TableCell>
          </TableRow>)}</TableBody>
        </Table>
      </div>}
      <div className="flex items-center justify-end gap-2 text-xs"><span>{cursors.length}</span><Button type="button" variant="outline" size="icon-sm" title={label('previous')} aria-label={label('previous')} disabled={loading || error || cursors.length <= 1} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => history.slice(0, -1)); }}><ChevronLeft /></Button><Button type="button" variant="outline" size="icon-sm" title={label('next')} aria-label={label('next')} disabled={loading || error || !hasMore || nextCursor <= 0} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => [...history, nextCursor]); }}><ChevronRight /></Button></div>
    </div>
  </section>;
}
