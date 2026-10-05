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
import { Tooltip, TooltipContent, TooltipTrigger } from '@gooseforum/ui/components/tooltip';
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
    <header><h2 className="text-xl font-semibold">{text.title}</h2></header>
    <div className="flex flex-wrap items-end gap-2 [&>[data-slot=field]]:w-[calc(50%-0.25rem)] sm:[&>[data-slot=field]]:w-36 [&>[data-slot=field]]:gap-1.5 [&>[data-slot=field]:nth-child(4)]:w-full sm:[&>[data-slot=field]:nth-child(4)]:w-64">
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-result">{text.result}</FieldLabel><Select value={result || 'all'} onValueChange={(value) => { setResult(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="auth-log-result" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'success', 'failure', 'challenge'].map((value) => <SelectItem key={value} value={value}>{option(value)}</SelectItem>)}</SelectContent></Select></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-method">{text.method}</FieldLabel><Select value={method || 'all'} onValueChange={(value) => { setMethod(value === 'all' ? '' : value); reset(); }}><SelectTrigger id="auth-log-method" className="w-full"><SelectValue /></SelectTrigger><SelectContent position="popper" align="start" className="w-(--radix-select-trigger-width) min-w-0">{['all', 'password', 'oauth', 'mfa', 'operator'].map((value) => <SelectItem key={value} value={value}>{option(value)}</SelectItem>)}</SelectContent></Select></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-user">{text.user}</FieldLabel><Input id="auth-log-user" type="number" min="1" value={userId} onChange={(e) => { setUserId(e.target.value); reset(); }} /></Field>
      <Field className="min-w-0"><FieldLabel htmlFor="auth-log-range">{text.dateRange}</FieldLabel><DateRangePicker id="auth-log-range" locale={locale} value={range} placeholder={text.rangePlaceholder} clearLabel={text.clearRange} onChange={(value) => { setRange(value); reset(); }} /></Field>
      <Tooltip><TooltipTrigger asChild><Button type="button" variant="outline" size="icon-sm" className="ml-auto mb-0.5" aria-label={text.refresh} disabled={loading} onClick={reset}><RefreshCw className={loading ? 'animate-spin motion-reduce:animate-none' : undefined} /></Button></TooltipTrigger><TooltipContent>{text.refresh}</TooltipContent></Tooltip>
    </div>
    {error ? <Alert variant="destructive"><AlertDescription>{text.error}</AlertDescription></Alert> : null}
    <section aria-busy={loading} className="min-w-0 overflow-hidden rounded-md border bg-background">
    {loading && !rows.length ? <div role="status" className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground"><Spinner role="presentation" aria-hidden="true" />{text.loading}</div> : !rows.length ? <Empty className="min-h-32 rounded-none border-0"><EmptyHeader><EmptyTitle>{error ? text.error : text.empty}</EmptyTitle></EmptyHeader></Empty> : <>
      {loading ? <span role="status" className="sr-only">{text.loading}</span> : null}
      <Table className="min-w-[64rem] table-fixed" containerProps={{ tabIndex: 0, role: 'region', 'aria-label': text.title, className: 'focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2' }}>
        <TableHeader className="bg-muted/30"><TableRow className="[&>th]:h-8"><TableHead className="w-44">{text.time}</TableHead><TableHead className="w-56">{text.event}</TableHead><TableHead className="w-20">{text.user}</TableHead><TableHead className="w-36">{text.method}</TableHead><TableHead className="w-24">{text.result}</TableHead><TableHead>{text.device}</TableHead></TableRow></TableHeader>
        <TableBody>{rows.map((row) => <TableRow key={row.id} className="h-14 [&>td]:py-2">
          <TableCell className="text-xs text-muted-foreground"><time dateTime={row.createdAt}>{new Date(row.createdAt).toLocaleString(locale, { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })}</time></TableCell>
          <TableCell><LogText value={dictionary.authLogActions[row.action as keyof typeof dictionary.authLogActions] || row.action} className="font-medium" />{row.reason && <LogText value={`${text.reason}: ${row.reason}`} className="mt-1 text-xs text-muted-foreground" />}</TableCell>
          <TableCell className="font-mono text-xs">{row.userId || '-'}</TableCell>
          <TableCell><LogText value={row.oauthProvider || option(row.authMethod || 'all')} /></TableCell>
          <TableCell><Badge variant={row.result === 'success' ? 'secondary' : 'outline'}>{option(row.result || 'success')}</Badge></TableCell>
          <TableCell className="text-xs"><LogText value={row.clientIp || '-'} className="font-mono" />{row.userAgent && <LogText value={row.userAgent} className="mt-1 text-muted-foreground" />}</TableCell>
        </TableRow>)}</TableBody>
      </Table>
    </>}
    </section>
    <footer className="flex items-center justify-between gap-3 text-xs"><span className="text-muted-foreground">{loading || error ? '-' : <>{rows.length ? (cursors.length - 1) * 20 + 1 : 0}-{rows.length ? (cursors.length - 1) * 20 + rows.length : 0}</>}</span><nav aria-label={text.title} className="flex items-center gap-2"><span className="min-w-8 text-center tabular-nums">{cursors.length}</span><Button type="button" variant="outline" size="icon-sm" title={text.previous} aria-label={text.previous} disabled={loading || error || cursors.length <= 1} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => history.slice(0, -1)); }}><ChevronLeft /></Button><Button type="button" variant="outline" size="icon-sm" title={text.next} aria-label={text.next} disabled={loading || error || !hasMore || nextCursor <= 0} onClick={() => { requestEpoch.current++; setLoading(true); setCursors((history) => [...history, nextCursor]); }}><ChevronRight /></Button></nav></footer>
  </AdminPage>;
}

function LogText({ value, className = '' }: { value: string; className?: string }) {
  return <Tooltip><TooltipTrigger asChild><p tabIndex={0} className={`truncate leading-4 ${className}`}>{value}</p></TooltipTrigger><TooltipContent className="max-w-sm break-words">{value}</TooltipContent></Tooltip>;
}
