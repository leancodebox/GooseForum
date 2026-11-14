import { useCallback, useEffect, useMemo, useState } from 'react'
import type { GooseAdminApi, RegistrationInvitation } from '@gooseforum/client'
import { AdminPage } from '../components/admin-page'
import { errorMessage } from '../error-message'
import { useLatestRequest } from '@gooseforum/runtime'
import type { AssetTextKey } from '../assets-i18n'
import { Badge } from '@gooseforum/ui/components/badge'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { DatePicker } from '@gooseforum/ui/components/date-picker'
import { Input } from '@gooseforum/ui/components/input'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { ChevronLeft, ChevronRight, Copy, Plus, RefreshCw, ShieldOff } from 'lucide-react'
import { toast } from '@gooseforum/ui/components/sonner'

type Text = (key: AssetTextKey) => string

type InvitationStatus = 'active' | 'used' | 'expired' | 'revoked'

function statusOf(item: RegistrationInvitation): { kind: InvitationStatus; variant: 'default' | 'secondary' | 'outline' | 'destructive' } {
  if (item.revokedAt) return { kind: 'revoked', variant: 'destructive' }
  if (item.expiresAt && new Date(item.expiresAt).getTime() <= Date.now()) return { kind: 'expired', variant: 'outline' }
  if (item.usedCount >= item.maxUses) return { kind: 'used', variant: 'secondary' }
  return { kind: 'active', variant: 'default' }
}

function formatDate(value: string | null | undefined, text: Text) {
  return value ? new Date(value).toLocaleString() : text('never')
}

function emailStatusKey(status: number): AssetTextKey {
  switch (status) {
    case 1: return 'emailSending'
    case 2: return 'emailSent'
    case 3: return 'emailFailed'
    case 4: return 'emailRetrying'
    default: return 'emailQueued'
  }
}

export function RegistrationInvitationsPage({ api, text, locale = 'en' }: { api: GooseAdminApi; text: Text; locale?: string }) {
  const [items, setItems] = useState<RegistrationInvitation[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [maxUses, setMaxUses] = useState(1)
  const [expiresAt, setExpiresAt] = useState('')
  const [email, setEmail] = useState('')
  const [saving, setSaving] = useState(false)
  const [newCode, setNewCode] = useState('')
  const [revokeTarget, setRevokeTarget] = useState<RegistrationInvitation | null>(null)

  const beginRequest = useLatestRequest()
  const load = useCallback(async () => {
    const isCurrent = beginRequest()
    setLoading(true)
    setError('')
    try {
      const result = await api.users.registrationInvitations({ page, pageSize })
      if (!isCurrent()) return
      setItems(result.list || [])
      setTotal(result.total || 0)
      setPageSize(result.pageSize || result.size || pageSize)
    } catch (reason) {
      if (isCurrent()) setError(errorMessage(reason, text('loadFailed')))
    } finally {
      if (isCurrent()) setLoading(false)
    }
  }, [api, beginRequest, page, pageSize, text])
  useEffect(() => { void load() }, [load])

  const activeCount = useMemo(() => items.filter((item) => statusOf(item).kind === 'active').length, [items])
  async function create() {
    setSaving(true)
    try {
      const input: { maxUses: number; expiresAt: string | null; email?: string } = { maxUses: Math.max(1, Math.trunc(maxUses)), expiresAt: expiresAt ? new Date(expiresAt).toISOString() : null }
      if (email.trim()) input.email = email.trim()
      const result = await api.users.createRegistrationInvitation(input)
      setNewCode(result.code)
      setCreateOpen(false)
      setMaxUses(1)
      setExpiresAt('')
      setEmail('')
      if (page === 1) void load()
      else setPage(1)
      toast.success(result.emailQueued ? text('invitationEmailQueued') : text('invitationSaved'))
    } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  async function revoke() {
    if (!revokeTarget) return
    setSaving(true)
    try { await api.users.revokeRegistrationInvitation(revokeTarget.id); setRevokeTarget(null); void load(); toast.success(text('invitationRevoked')) } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  async function retryEmail(item: RegistrationInvitation) {
    setSaving(true)
    try { await api.users.retryRegistrationInvitationEmail(item.id); void load(); toast.success(text('invitationEmailRetryQueued')) } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  async function copyCode() {
    if (!newCode) return
    await navigator.clipboard?.writeText(newCode)
    toast.success(text('copied'))
  }
  async function copyRegistrationLink() {
    if (!newCode) return
    const link = `${window.location.origin}/login?register=true&invitation=${encodeURIComponent(newCode)}`
    await navigator.clipboard?.writeText(link)
    toast.success(text('invitationLinkCopied'))
  }

  return <AdminPage>
    <header className="flex items-center justify-between gap-3">
      <div><h2 className="text-lg font-semibold">{text('registrationInvitations')}</h2><p className="text-xs text-muted-foreground">{text('registrationInvitationsHint')}</p></div>
      <Button size="sm" onClick={() => setCreateOpen(true)}><Plus data-icon="inline-start" />{text('newInvitation')}</Button>
    </header>
    <div className="flex items-center justify-between"><Badge variant="secondary">{text('activeOnPage')}: {activeCount}</Badge><Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button></div>
    {error ? <div className="rounded-lg border border-destructive/30 p-3 text-destructive">{error}</div> : null}
    {loading && !items.length ? <div className="flex min-h-32 items-center justify-center"><Spinner /></div> : !items.length ? <div className="rounded-lg border p-6 text-center text-sm text-muted-foreground">{text('empty')}</div> : <div className="grid gap-2">{items.map((item) => { const status = statusOf(item); return <article key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-background p-3"><div className="min-w-0"><div className="flex items-center gap-2"><code className="font-mono text-sm">{item.prefix}…</code><Badge variant={status.variant}>{text(status.kind)}</Badge></div><p className="mt-1 text-xs text-muted-foreground">{text('used')}: {item.usedCount}/{item.maxUses} · {text('expiresAt')}: {formatDate(item.expiresAt, text)}{item.emailTaskStatus !== undefined ? <> · {text('emailStatus')}: {text(emailStatusKey(item.emailTaskStatus))}</> : null}</p></div><div className="flex gap-2">{item.emailTaskStatus === 3 ? <Button variant="outline" size="sm" disabled={saving} onClick={() => void retryEmail(item)}><RefreshCw data-icon="inline-start" /><span>{text('retryInvitationEmail')}</span></Button> : null}{status.kind === 'active' ? <Button variant="outline" size="sm" onClick={() => setRevokeTarget(item)}><ShieldOff data-icon="inline-start" /><span>{text('revokeInvitation')}</span></Button> : null}</div></article> })}</div>}
    <footer className="flex items-center justify-between gap-3 text-xs text-muted-foreground"><span>{total ? (page - 1) * pageSize + 1 : 0}-{Math.min(page * pageSize, total)} / {total}</span><nav aria-label={text('page')} className="flex items-center gap-2"><Button variant="outline" size="icon-sm" aria-label={text('previous')} disabled={page <= 1 || loading || saving} onClick={() => setPage(page - 1)}><ChevronLeft /></Button><span>{text('page')} {page}/{Math.max(1, Math.ceil(total / pageSize))}</span><Button variant="outline" size="icon-sm" aria-label={text('next')} disabled={page >= Math.max(1, Math.ceil(total / pageSize)) || loading || saving} onClick={() => setPage(page + 1)}><ChevronRight /></Button></nav></footer>
    <Dialog open={createOpen} onOpenChange={(open) => !saving && setCreateOpen(open)}><DialogContent><DialogHeader><DialogTitle>{text('newInvitation')}</DialogTitle></DialogHeader><div className="grid gap-4"><Field><FieldLabel htmlFor="invitation-max-uses">{text('maxUses')}</FieldLabel><Input id="invitation-max-uses" type="number" min={1} max={100000} value={maxUses} onChange={(event) => setMaxUses(Number(event.target.value))} /></Field><Field><FieldLabel htmlFor="invitation-expires-at">{text('expiresAt')}</FieldLabel><DatePicker id="invitation-expires-at" withTime locale={locale} value={expiresAt} onChange={setExpiresAt} placeholder={text('never')} clearLabel={text('dateClear')} timeLabel={text('dateTime')} /></Field><Field><FieldLabel htmlFor="invitation-email">{text('invitationEmail')}</FieldLabel><Input id="invitation-email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder={text('invitationEmailHint')} /></Field></div><DialogFooter><Button variant="outline" onClick={() => setCreateOpen(false)}>{text('cancel')}</Button><Button disabled={saving} onClick={() => void create()}>{saving ? <Spinner data-icon="inline-start" /> : null}{text('createInvitation')}</Button></DialogFooter></DialogContent></Dialog>
    <Dialog open={Boolean(newCode)} onOpenChange={(open) => !open && setNewCode('')}><DialogContent><DialogHeader><DialogTitle>{text('invitationCreated')}</DialogTitle><DialogDescription>{text('invitationCodeHint')}</DialogDescription></DialogHeader><div className="flex flex-wrap gap-2"><Input readOnly value={newCode} /><Button variant="outline" onClick={() => void copyCode()}><Copy data-icon="inline-start" />{text('copyCode')}</Button><Button variant="outline" onClick={() => void copyRegistrationLink()}><Copy data-icon="inline-start" />{text('copyInvitationLink')}</Button></div></DialogContent></Dialog>
    <Dialog open={Boolean(revokeTarget)} onOpenChange={(open) => !open && setRevokeTarget(null)}><DialogContent><DialogHeader><DialogTitle>{text('revokeInvitation')}</DialogTitle><DialogDescription>{text('revokeInvitationConfirm')}</DialogDescription></DialogHeader><DialogFooter><Button variant="outline" onClick={() => setRevokeTarget(null)}>{text('cancel')}</Button><Button variant="destructive" disabled={saving} onClick={() => void revoke()}>{text('revokeInvitation')}</Button></DialogFooter></DialogContent></Dialog>
  </AdminPage>
}
