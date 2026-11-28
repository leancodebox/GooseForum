import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { useLatestRequest } from '@gooseforum/runtime'
import type { AdminBadge, AdminBadgeRecipient, AdminUser, GooseAdminApi } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Textarea } from '@gooseforum/ui/components/textarea'
import { toast } from '@gooseforum/ui/components/sonner'
import { Check, ChevronLeft, ChevronRight, Plus, Search } from 'lucide-react'
import { errorMessage } from '../error-message'
import type { BadgeText } from './badge-options'

export function BadgeRecipients({ badge, api, text, onClose, onChanged }: {
  badge: AdminBadge; api: GooseAdminApi; text: BadgeText; onClose(): void; onChanged(): Promise<void>
}) {
  const [items, setItems] = useState<AdminBadgeRecipient[]>([])
  const [total, setTotal] = useState<number | undefined>(undefined)
  const [hasNext, setHasNext] = useState(false)
  const [page, setPage] = useState(1)
  const [query, setQuery] = useState('')
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [action, setAction] = useState<{ kind: 'grant' | 'revoke'; recipient?: AdminBadgeRecipient } | null>(null)
  const beginRequest = useLatestRequest()
  const pageSize = 20
  const load = useCallback(() => {
    const isCurrent = beginRequest()
    return api.assets.badgeRecipients({ code: badge.code, page, pageSize, search }).then(result => {
      if (!isCurrent()) return
      setError('')
      setItems(result.list || [])
      setTotal(search ? undefined : result.total)
      setHasNext(result.hasNext)
      if (page > 1 && !(result.list || []).length) setPage(page - 1)
    }).catch(reason => {
      if (isCurrent()) setError(errorMessage(reason, text('loadFailed')))
    }).finally(() => { if (isCurrent()) setLoading(false) })
  }, [api, badge.code, beginRequest, page, search, text])
  useEffect(() => { void load() }, [load])
  async function changed() { setLoading(true); await Promise.all([load(), onChanged()]) }

  return <Dialog open onOpenChange={open => { if (!open && !action) onClose() }}><DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden sm:max-w-3xl">
    <DialogHeader className="pr-6"><DialogTitle>{badge.name} · {text('recipients')}</DialogTitle><DialogDescription>{text('badgeRecipientsHint')}</DialogDescription></DialogHeader>
    <div className="flex flex-wrap items-center gap-2"><form className="flex min-w-0 flex-1 gap-2" onSubmit={event => { event.preventDefault(); setLoading(true); if (page === 1 && query.trim() === search) void load(); else { setPage(1); setSearch(query.trim()) } }}><Input aria-label={text('recipientSearch')} placeholder={text('recipientSearch')} value={query} onChange={event => setQuery(event.target.value)} /><Button variant="outline" type="submit" size="icon" aria-label={text('search')}><Search /></Button></form><Button disabled={!badge.isEnabled || badge.grantMode !== 'manual'} onClick={() => setAction({ kind: 'grant' })}><Plus data-icon="inline-start" />{text('grantBadge')}</Button></div>
    {!badge.isEnabled || badge.grantMode !== 'manual' ? <p className="text-xs text-muted-foreground">{text(!badge.isEnabled ? 'disabledGrantHint' : 'automaticGrantHint')}</p> : null}
    <div className="min-h-24 min-w-0 overflow-y-auto rounded-lg border" aria-busy={loading}>
      {loading ? <div className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground"><Spinner />{text('loading')}</div> : error ? <div role="alert" className="p-4 text-sm text-destructive">{error}<Button variant="outline" className="ml-2" size="sm" onClick={() => { setLoading(true); void load() }}>{text('retry')}</Button></div> : !items.length ? <p className="p-6 text-center text-sm text-muted-foreground">{text(search ? 'noMatchingMembers' : 'noRecipients')}</p> : <div className="divide-y">{items.map(item => <article key={item.userId} className="flex items-start gap-3 p-3">
        <div className="min-w-0 flex-1"><div className="flex flex-wrap items-center gap-x-2 gap-y-1"><a className="break-words text-sm font-medium hover:underline" href={`/u/${item.userId}`} target="_blank" rel="noreferrer">{item.username || `#${item.userId}`}</a><span className="text-xs text-muted-foreground">#{item.userId} · {text(item.source === 'auto' ? 'auto' : item.source === 'migration' ? 'migrationSource' : 'manual')}</span></div><p className="mt-1 break-words text-xs text-muted-foreground">{text('grantedAt')} {formatDate(item.grantedAt)} · {text('grantedBy')} {item.grantedByName || (item.grantedBy ? `#${item.grantedBy}` : text('system'))}</p><p className="mt-1 whitespace-pre-wrap break-words text-sm">{item.reason || text('noGrantReason')}</p></div>
        <Button size="sm" variant="outline" onClick={() => setAction({ kind: 'revoke', recipient: item })}>{text('revokeBadge')}</Button>
      </article>)}</div>}
    </div>
    <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground"><span>{!search && total !== undefined ? text('recipientCount', { count: total }) : null}</span><div className="flex items-center gap-2"><Button variant="outline" size="icon-sm" aria-label={text('previous')} disabled={loading || page <= 1} onClick={() => { setLoading(true); setPage(page - 1) }}><ChevronLeft /></Button><span>{text('recipientPage', { page })}</span><Button variant="outline" size="icon-sm" aria-label={text('next')} disabled={loading || !hasNext} onClick={() => { setLoading(true); setPage(page + 1) }}><ChevronRight /></Button></div></div>
    {action ? <RecipientAction badge={badge} action={action.kind} recipient={action.recipient} api={api} text={text} onClose={() => setAction(null)} onSaved={changed} /> : null}
  </DialogContent></Dialog>
}

function RecipientAction({ badge, action, recipient, api, text, onClose, onSaved }: {
  badge: AdminBadge; action: 'grant' | 'revoke'; recipient?: AdminBadgeRecipient; api: GooseAdminApi; text: BadgeText; onClose(): void; onSaved(): Promise<void>
}) {
  const [query, setQuery] = useState('')
  const [users, setUsers] = useState<AdminUser[]>([])
  const [selected, setSelected] = useState<AdminUser | null>(null)
  const [searched, setSearched] = useState(false)
  const [searching, setSearching] = useState(false)
  const [reason, setReason] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const beginSearch = useLatestRequest()
  async function findUsers() {
    const trimmed = query.trim()
    const isCurrent = beginSearch()
    setUsers([])
    setSelected(null)
    if (!trimmed) { setSearched(false); return }
    setSearching(true)
    setError('')
    try {
      const result = await api.users.list({ username: trimmed, page: 1, pageSize: 20 })
      if (isCurrent()) { setUsers(result.list || []); setSearched(true) }
    } catch (reason) { if (isCurrent()) setError(errorMessage(reason, text('loadFailed'))) }
    finally { if (isCurrent()) setSearching(false) }
  }
  async function submit(event: FormEvent) {
    event.preventDefault()
    if (saving) return
    const userId = action === 'revoke' ? recipient?.userId : selected?.userId
    if (!userId) { setError(text('selectMemberFirst')); return }
    if (!reason.trim()) { setError(text('badgeReasonRequired')); return }
    setSaving(true)
    setError('')
    try {
      await api.assets.updateBadgeRecipient({ code: badge.code, userId, action, reason: reason.trim() })
      toast.success(text(action === 'grant' ? 'badgeGranted' : 'badgeRevoked'))
      await onSaved()
      onClose()
    } catch (reason) { setError(errorMessage(reason, text('saveFailed'))) }
    finally { setSaving(false) }
  }
  return <Dialog open onOpenChange={open => { if (!open && !saving) onClose() }}><DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden sm:max-w-lg" showCloseButton={!saving}><DialogHeader><DialogTitle>{text(action === 'grant' ? 'grantBadge' : 'revokeBadge')}</DialogTitle><DialogDescription>{badge.name}{recipient ? ` · ${recipient.username || `#${recipient.userId}`}` : ''}</DialogDescription></DialogHeader>
    <form id="badge-recipient-action" onSubmit={submit} className="-m-1 min-h-0 space-y-4 overflow-y-auto scroll-p-1 p-1">
      {action === 'grant' ? <Field><FieldLabel htmlFor="badge-user-search">{text('selectMember')}</FieldLabel><div className="flex gap-2"><Input id="badge-user-search" placeholder={text('memberSearchHint')} disabled={saving} value={query} onChange={event => { beginSearch(); setSearching(false); setQuery(event.target.value); setSelected(null); setSearched(false); setUsers([]) }} onKeyDown={event => { if (event.key === 'Enter') { event.preventDefault(); void findUsers() } }} /><Button variant="outline" type="button" disabled={saving || searching || !query.trim()} onClick={() => void findUsers()}>{searching ? <Spinner /> : <Search />}{text('search')}</Button></div>{users.length ? <div className="max-h-48 space-y-1 overflow-y-auto rounded-lg border p-1" role="group" aria-label={text('selectMember')}>{users.map(user => <Button key={user.userId} type="button" variant={selected?.userId === user.userId ? 'secondary' : 'ghost'} className="h-auto w-full justify-start whitespace-normal text-left" aria-pressed={selected?.userId === user.userId} disabled={saving} onClick={() => setSelected(user)}><span className="min-w-0 flex-1 break-words">{user.username} <span className="text-xs text-muted-foreground">#{user.userId}</span></span>{selected?.userId === user.userId ? <Check /> : null}</Button>)}</div> : searched && !searching ? <p className="text-xs text-muted-foreground">{text('noMatchingMembers')}</p> : null}{users.length === 20 ? <p className="text-xs text-muted-foreground">{text('refineMemberSearch')}</p> : null}</Field> : <p className="rounded-lg bg-muted p-3 text-sm">{text('revokeBadgeHint')}</p>}
      <Field><FieldLabel htmlFor="badge-action-reason">{text(action === 'grant' ? 'grantReason' : 'revokeReason')}</FieldLabel><Textarea id="badge-action-reason" disabled={saving} maxLength={255} value={reason} onChange={event => setReason(event.target.value)} rows={3} /></Field>
      {error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}
    </form><DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button form="badge-recipient-action" type="submit" variant={action === 'revoke' ? 'destructive' : 'default'} disabled={saving || searching}>{saving ? <Spinner data-icon="inline-start" /> : null}{text(action === 'grant' ? 'grantBadge' : 'revokeBadge')}</Button></DialogFooter>
  </DialogContent></Dialog>
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
