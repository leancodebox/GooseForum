import { useCallback, useEffect, useState } from 'react'
import type { AdminOptRecord, GooseAdminApi } from '@gooseforum/client'
import type { AuthLocale } from '@gooseforum/react/i18n/auth'
import { Badge } from '@gooseforum/react/components/ui/badge'
import { Button } from '@gooseforum/react/components/ui/button'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/react/components/ui/empty'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/react/components/ui/select'
import { Spinner } from '@gooseforum/react/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/react/components/ui/table'
import { ChevronLeft, ChevronRight, ListChecks, RefreshCw } from 'lucide-react'
import { formatAuditMessage, type AuditTextKey } from '../audit-i18n'

type Text = (key: AuditTextKey) => string
const optTypes: Record<number, AuditTextKey> = { 0: 'editUser', 1: 'editTopic', 2: 'editCategory' }
const targetTypes: Record<number, AuditTextKey> = { 0: 'system', 1: 'user', 2: 'topic', 3: 'docProject', 4: 'docVersion', 5: 'docContent', 6: 'category' }

export function OptRecordsManagementPage({ api, text, locale }: { api: GooseAdminApi; text: Text; locale: AuthLocale }) {
  const [items, setItems] = useState<AdminOptRecord[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const load = useCallback(async () => {
    setLoading(true); setError('')
    try {
      const result = await api.audit.records({ page, pageSize })
      setItems(result.list || []); setTotal(result.total || 0)
      setPage((result.page ?? page - 1) + 1)
      setPageSize(result.pageSize || result.size || pageSize)
    } catch (reason) {
      setError(reason instanceof Error && reason.message ? reason.message : text('loadFailed'))
    } finally { setLoading(false) }
  }, [api, page, pageSize, text])

  useEffect(() => { void load() }, [load])

  return <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
    <header className="flex items-center justify-between gap-3"><div><h2 className="text-lg font-semibold">{text('title')}</h2><p className="text-xs text-muted-foreground">{text('description')}</p></div><Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button></header>
    <section className="overflow-hidden rounded-lg border bg-background">
      {loading && !items.length ? <LogEmpty icon={<Spinner />} title={text('loading')} /> : error ? <LogEmpty icon={<ListChecks />} title={error} /> : !items.length ? <LogEmpty icon={<ListChecks />} title={text('empty')} /> : <Table className="min-w-240 table-fixed"><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8 w-14">ID</TableHead><TableHead className="h-8 w-24">{text('operator')}</TableHead><TableHead className="h-8 w-32">{text('operation')}</TableHead><TableHead className="h-8 w-32">{text('targetType')}</TableHead><TableHead className="h-8 w-28">{text('targetId')}</TableHead><TableHead className="h-8">{text('details')}</TableHead><TableHead className="h-8 w-44">{text('time')}</TableHead></TableRow></TableHeader><TableBody>{items.map((item) => <TableRow key={item.id}><TableCell className="py-2 font-mono text-xs text-muted-foreground">{item.id}</TableCell><TableCell className="py-2 font-mono text-xs">{item.optUserId || '—'}</TableCell><TableCell className="py-2"><Badge variant="secondary">{text(optTypes[item.optType] || 'unknown')} {optTypes[item.optType] ? null : item.optType}</Badge></TableCell><TableCell className="py-2 text-muted-foreground">{text(targetTypes[item.targetType] || 'unknown')} {targetTypes[item.targetType] ? null : item.targetType}</TableCell><TableCell className="py-2 font-mono text-xs text-muted-foreground">{item.targetId || '—'}</TableCell><TableCell className="py-2"><div className="line-clamp-2">{optInfo(item, locale)}</div></TableCell><TableCell className="whitespace-nowrap py-2 text-muted-foreground">{formatTime(item.createdAt, locale)}</TableCell></TableRow>)}</TableBody></Table>}
      <footer className="flex items-center justify-between gap-3 border-t bg-muted/10 px-3 py-2 text-sm text-muted-foreground"><span>{total}</span><div className="flex items-center gap-1.5"><Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1) }}><SelectTrigger className="w-20"><SelectValue /></SelectTrigger><SelectContent><SelectGroup>{[10, 20, 30, 50].map((size) => <SelectItem key={size} value={String(size)}>{size}</SelectItem>)}</SelectGroup></SelectContent></Select><Button variant="outline" size="icon-sm" disabled={page <= 1} onClick={() => setPage(page - 1)}><ChevronLeft /></Button><span>{text('page')} {page}/{totalPages}</span><Button variant="outline" size="icon-sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}><ChevronRight /></Button></div></footer>
    </section>
  </main>
}

function optInfo(item: AdminOptRecord, locale: AuthLocale) {
  const payload = item.optInfoPayload || parsePayload(item.optInfo)
  if (!payload?.messageCode) return item.optInfo || '—'
  return formatAuditMessage(locale, payload.messageCode, payload.params || {}, item.optInfo)
}

function parsePayload(value: string) {
  if (!value || value[0] !== '{') return undefined
  try {
    const parsed = JSON.parse(value) as { messageCode?: unknown; params?: unknown }
    if (typeof parsed.messageCode !== 'string') return undefined
    return { messageCode: parsed.messageCode, params: parsed.params && typeof parsed.params === 'object' && !Array.isArray(parsed.params) ? parsed.params as Record<string, unknown> : {} }
  } catch { return undefined }
}

function formatTime(value: string, locale: AuthLocale) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString(locale === 'zh' ? 'zh-CN' : locale, { hour12: false })
}

function LogEmpty({ icon, title }: { icon: React.ReactNode; title: string }) { return <Empty className="min-h-32 border-0"><EmptyHeader><EmptyMedia variant="icon">{icon}</EmptyMedia><EmptyTitle>{title}</EmptyTitle></EmptyHeader></Empty> }
