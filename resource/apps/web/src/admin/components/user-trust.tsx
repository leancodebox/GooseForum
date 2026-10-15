import { useEffect, useRef, useState } from 'react'
import { trustLevelName, trustLevels, type GooseAdminApi, type TrustProgress, type TrustHistory } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { Switch } from '@gooseforum/ui/components/switch'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Save, RefreshCw, ChevronDown } from 'lucide-react'

export function UserTrust({ api, userId, locale }: { api: GooseAdminApi; userId: number; locale: string }) {
  const zh = locale.startsWith('zh')
  const [progress, setProgress] = useState<TrustProgress>()
  const [history, setHistory] = useState<TrustHistory[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [level, setLevel] = useState(0)
  const [locked, setLocked] = useState(false)
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [reload, setReload] = useState(0)
  const inFlight = useRef(false)
  useEffect(() => {
    let active = true
    void Promise.all([api.trust.user(userId), api.trust.history(userId)]).then(([p, h]) => { if (active) { setProgress(p); setLevel(p.level); setLocked(p.locked); setHistory(h); setHasMore(h.length === 30) } }).catch(e => { if (active) setError(String(e)) }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [api, userId, reload])
  async function save() {
    if (!progress || !reason.trim() || inFlight.current) return
    inFlight.current = true; setBusy(true); setError('')
    try { const p = await api.trust.setUser({ userId, level, locked, expectedVersion: progress.version, reason }); setProgress(p); setLevel(p.level); setLocked(p.locked); setReason(''); setHistory(await api.trust.history(userId)) }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); setReload(v => v + 1) }
    finally { inFlight.current = false; setBusy(false) }
  }
  async function more() {
    if (busy || !history.length) return
    setBusy(true)
    try { const rows = await api.trust.history(userId, history[history.length - 1].id); setHistory(current => [...current, ...rows]); setHasMore(rows.length === 30) }
    catch (e) { setError(String(e)) }
    finally { setBusy(false) }
  }
  return <section className="space-y-4">
    {error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}
    {loading ? <Spinner /> : progress ? <>
      <div className="text-sm text-muted-foreground">{zh ? '访问日 / 不同主题' : 'Visit days / distinct topics'}: {progress.visitDays} / {progress.topicsVisited}</div>
      <Field><FieldLabel htmlFor="user-trust-level">{zh ? '等级' : 'Level'}</FieldLabel><Select value={String(level)} disabled={busy} onValueChange={value => setLevel(Number(value))}><SelectTrigger id="user-trust-level"><SelectValue>{trustLevelName(level, locale)}</SelectValue></SelectTrigger><SelectContent>{trustLevels.filter(value => value >= progress.level).map(value => <SelectItem key={value} value={String(value)}>{trustLevelName(value, locale)}</SelectItem>)}</SelectContent></Select></Field>
      <Field orientation="horizontal"><FieldLabel htmlFor="user-trust-lock">{zh ? '锁定自动晋级' : 'Lock automatic promotion'}</FieldLabel><Switch id="user-trust-lock" disabled={busy} checked={locked} onCheckedChange={setLocked} /></Field>
      <Field><FieldLabel htmlFor="user-trust-reason">{zh ? '调整原因' : 'Reason'}</FieldLabel><Input id="user-trust-reason" disabled={busy} maxLength={500} value={reason} onChange={event => setReason(event.target.value)} onKeyDown={event => { if (event.key === 'Enter') { event.preventDefault(); void save() } }} /></Field>
      <Button type="button" size="sm" disabled={busy || !reason.trim() || level === progress.level && locked === progress.locked} onClick={() => void save()}>{busy ? <Spinner /> : <Save />}{zh ? '保存等级' : 'Save level'}</Button>
      <div className="space-y-2 border-t pt-3"><h4 className="text-sm font-medium">{zh ? '等级历史' : 'Level history'}</h4>{history.length ? history.map(row => <div key={row.id} className="border-b pb-2 text-xs"><div>{trustLevelName(row.fromLevel, locale)} → {trustLevelName(row.toLevel, locale)} · {row.toLocked ? (zh ? '已锁定' : 'Locked') : (zh ? '未锁定' : 'Unlocked')}</div><div className="mt-1 break-words text-muted-foreground">{new Date(row.createdAt).toLocaleString(locale)} · {row.source} {row.operatorId ? `#${row.operatorId}` : ''} · {row.reason}</div></div>) : <p className="text-xs text-muted-foreground">{zh ? '暂无历史记录' : 'No history'}</p>}{hasMore ? <Button type="button" variant="outline" size="sm" disabled={busy} onClick={() => void more()}><ChevronDown />{zh ? '更多' : 'More'}</Button> : null}</div>
    </> : <Button type="button" size="sm" variant="outline" onClick={() => { setLoading(true); setReload(v => v + 1) }}><RefreshCw />{zh ? '重试' : 'Retry'}</Button>}
  </section>
}
