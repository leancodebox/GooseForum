import { useEffect, useRef, useState } from 'react';
import type { AuthLogPayload, GooseAdminApi } from '@gooseforum/client';
import type { AuthLocale } from '@gooseforum/runtime/i18n/auth';
import { Button } from '@gooseforum/ui/components/button';
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert';
import { Badge } from '@gooseforum/ui/components/badge';
import { Empty, EmptyHeader, EmptyTitle } from '@gooseforum/ui/components/empty';
import { Spinner } from '@gooseforum/ui/components/spinner';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table';
import { Input } from '@gooseforum/ui/components/input';
import { DateRangePicker } from '@gooseforum/ui/components/date-range-picker';
import { Field, FieldLabel } from '@gooseforum/ui/components/field';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select';
import { ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react';
import { AdminPage } from '../components/admin-page';
import { getAdminDictionary } from '../translation-loader';

export function AuthLogsPage({ api, locale }: { api: GooseAdminApi; locale: AuthLocale }) {
  const dictionary = getAdminDictionary<typeof import('../messages/en-audit').default>('audit', locale);
  const text = dictionary.authLogs;
  const [rows, setRows] = useState<AuthLogPayload[]>([]);
  const [cursors, setCursors] = useState([0]);
  const [nextCursor, setNextCursor] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [result, setResult] = useState('');
  const [method, setMethod] = useState('');
  const [userId, setUserId] = useState('');
  const [range, setRange] = useState({ from: '', to: '' });
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const requestEpoch = useRef(0);
  const [owner, setOwner] = useState(api);
  if (owner !== api) {
    setOwner(api); setCursors([0]); setRows([]); setNextCursor(0); setHasMore(false);
    setResult(''); setMethod(''); setUserId(''); setRange({ from: '', to: '' });
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
      return api.audit.authLogs({ cursor, pageSize: 20, result, method, userId: Number(userId) || undefined,
      since: range.from ? new Date(`${range.from}T00:00:00`).toISOString() : undefined,
      until: range.to ? new Date(`${range.to}T23:59:59.999`).toISOString() : undefined,
      });
    }).then((data) => { if (isCurrent() && data) { setRows(data.list); setNextCursor(data.nextCursor); setHasMore(data.hasMore); } })
      .catch(() => { if (isCurrent()) setError(true); })
      .finally(() => { if (isCurrent()) setLoading(false); });
    return () => { active = false; };
  }, [api, cursor, result, method, userId, range, refresh]);
  const option = (value: string) => text[(value || 'all') as keyof typeof text] || value;
  return <AdminPage className="min-w-0">
    <header className="flex items-center justify-between gap-3"><h2 className="text-lg font-semibold">{text.title}</h2><Button type="button" variant="outline" size="icon" title={text.refresh} aria-label={text.refresh} disabled={loading} onClick={reset}><RefreshCw /></Button></header>
    <div className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,14rem),1fr))] items-end gap-3 border-b pb-3">
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-result">{text.result}</FieldLabel><Select value={result || 'all'} onValueChange={(value) => { setResult(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="auth-log-result" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'success', 'failure', 'challenge'].map((value) => <SelectItem key={value} value={value}>{option(value)}</SelectItem>)}</SelectContent></Select></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-method">{text.method}</FieldLabel><Select value={method || 'all'} onValueChange={(value) => { setMethod(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="auth-log-method" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'password', 'oauth', 'mfa', 'operator'].map((value) => <SelectItem key={value} value={value}>{option(value)}</SelectItem>)}</SelectContent></Select></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-user">{text.user}</FieldLabel><Input id="auth-log-user" type="number" min="1" value={userId} onChange={(e) => { setUserId(e.target.value); reset(); }} /></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-range">{text.dateRange}</FieldLabel><DateRangePicker id="auth-log-range" locale={locale} value={range} placeholder={text.rangePlaceholder} clearLabel={text.clearRange} onChange={(value) => { setRange(value); reset(); }} /></Field>
    </div>
    {error ? <Alert><AlertDescription>{text.error}</AlertDescription></Alert> : loading ? <div role="status" className="flex items-center gap-2 py-4 text-sm text-muted-foreground"><Spinner role="presentation" aria-hidden="true" />{text.loading}</div> : !rows.length ? <Empty className="min-h-20 rounded-none p-4"><EmptyHeader><EmptyTitle className="font-normal tracking-normal text-muted-foreground">{text.empty}</EmptyTitle></EmptyHeader></Empty> : <div className="min-w-0 max-w-full border-y">
      <Table className="min-w-[48rem] table-fixed" containerProps={{ tabIndex: 0, role: 'region', 'aria-label': text.title, className: 'focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2' }}>
        <TableHeader><TableRow><TableHead className="w-40">{text.time}</TableHead><TableHead className="w-56">{text.event}</TableHead><TableHead className="w-20">{text.user}</TableHead><TableHead className="w-36">{text.method}</TableHead><TableHead className="w-24">{text.result}</TableHead><TableHead className="w-56">{text.device}</TableHead></TableRow></TableHeader>
        <TableBody>{rows.map((row) => <TableRow key={row.id}>
          <TableCell className="whitespace-normal text-xs"><time dateTime={row.createdAt}>{new Date(row.createdAt).toLocaleString(locale)}</time></TableCell>
          <TableCell className="whitespace-normal break-words"><span className="font-medium">{dictionary.authLogActions[row.action as keyof typeof dictionary.authLogActions] || row.action}</span>{row.reason && <p className="mt-1 text-xs text-muted-foreground">{text.reason}: {row.reason}</p>}</TableCell>
          <TableCell>{row.userId || '-'}</TableCell>
          <TableCell className="whitespace-normal break-words">{row.oauthProvider || option(row.authMethod || 'all')}</TableCell>
          <TableCell><Badge variant={row.result === 'success' ? 'secondary' : 'outline'}>{option(row.result || 'success')}</Badge></TableCell>
          <TableCell className="whitespace-normal break-all text-xs"><p>{row.clientIp || '-'}</p>{row.userAgent && <p className="mt-1 text-muted-foreground">{row.userAgent}</p>}</TableCell>
        </TableRow>)}</TableBody>
      </Table>
    </div>}
    <div className="flex items-center justify-end gap-2 text-xs"><span>{cursors.length}</span><Button type="button" variant="outline" size="icon-sm" title={text.previous} aria-label={text.previous} disabled={loading || error || cursors.length <= 1} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => history.slice(0, -1)); }}><ChevronLeft /></Button><Button type="button" variant="outline" size="icon-sm" title={text.next} aria-label={text.next} disabled={loading || error || !hasMore || nextCursor <= 0} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => [...history, nextCursor]); }}><ChevronRight /></Button></div>
  </AdminPage>;
}
