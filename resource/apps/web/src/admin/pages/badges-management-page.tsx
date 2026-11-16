import { useCallback, useEffect, useMemo, useState } from 'react'
import { useLatestRequest } from '@gooseforum/runtime'
import type { AdminBadge, GooseAdminApi } from '@gooseforum/client'
import { Badge } from '@gooseforum/ui/components/badge'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { ForumBadge } from '@gooseforum/ui/components/forum-badge'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { toast } from '@gooseforum/ui/components/sonner'
import { Edit3, Plus, RefreshCw, Search, Trash2, Users } from 'lucide-react'
import { AdminPage } from '../components/admin-page'
import { BadgeEditor } from '../components/badge-editor'
import { BadgeRecipients } from '../components/badge-recipients'
import { badgeLevelLabel, newBadge, type BadgeText } from '../components/badge-options'
import { errorMessage } from '../error-message'

export function BadgesManagementPage({ api, text, canManageRecipients = false }: { api: GooseAdminApi; text: BadgeText; canManageRecipients?: boolean }) {
  const [items, setItems] = useState<AdminBadge[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [type, setType] = useState('all')
  const [status, setStatus] = useState('all')
  const [editing, setEditing] = useState<AdminBadge | null>(null)
  const [members, setMembers] = useState<AdminBadge | null>(null)
  const [deleting, setDeleting] = useState<AdminBadge | null>(null)
  const beginRequest = useLatestRequest()
  const load = useCallback(() => {
    const isCurrent = beginRequest()
    return api.assets.badges().then(result => {
      if (isCurrent()) { setItems(result); setError('') }
    }).catch(reason => {
      if (isCurrent()) setError(errorMessage(reason, text('loadFailed')))
    }).finally(() => { if (isCurrent()) setLoading(false) })
  }, [api, beginRequest, text])
  useEffect(() => { void load() }, [load])
  const filtered = useMemo(() => items.filter(item => {
    const query = search.trim().toLocaleLowerCase()
    return (!query || `${item.name} ${item.description} ${item.code}`.toLocaleLowerCase().includes(query))
      && (type === 'all' || item.type === type)
      && (status === 'all' || item.isEnabled === (status === 'enabled'))
  }), [items, search, status, type])
  const systemCount = items.filter(item => item.type === 'system').length

  return <AdminPage>
    <header className="flex flex-wrap items-center justify-between gap-3"><div><h2 className="text-lg font-semibold">{text('badges')}</h2><p className="text-xs text-muted-foreground">{text('badgesHint')}</p></div><Button size="sm" onClick={() => setEditing({ ...newBadge })}><Plus data-icon="inline-start" />{text('newBadge')}</Button></header>
    <div className="flex flex-wrap items-center justify-between gap-2"><div className="flex gap-2"><Badge variant="secondary">{text('system')} {systemCount}</Badge><Badge variant="outline">{text('custom')} {items.length - systemCount}</Badge></div><Button variant="outline" size="sm" disabled={loading} onClick={() => { setLoading(true); void load() }}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button></div>
    <section className="overflow-hidden rounded-lg border bg-background" aria-label={text('badges')}>
      <div className="flex flex-wrap gap-2 border-b p-3"><div className="relative min-w-0 flex-[1_1_14rem]"><Search aria-hidden className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" /><Input aria-label={text('badgeSearch')} placeholder={text('badgeSearch')} value={search} onChange={event => setSearch(event.target.value)} className="pl-8" /></div><Select value={type} onValueChange={setType}><SelectTrigger aria-label={text('type')} className="min-w-28"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="all">{text('allTypes')}</SelectItem><SelectItem value="system">{text('system')}</SelectItem><SelectItem value="custom">{text('custom')}</SelectItem></SelectContent></Select><Select value={status} onValueChange={setStatus}><SelectTrigger aria-label={text('badgeStatus')} className="min-w-28"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="all">{text('allStatuses')}</SelectItem><SelectItem value="enabled">{text('enabled')}</SelectItem><SelectItem value="disabled">{text('disabled')}</SelectItem></SelectContent></Select></div>
      {error ? <div role="alert" className="p-4 text-sm text-destructive">{error}</div> : loading && !items.length ? <div className="flex min-h-48 items-center justify-center gap-2 text-muted-foreground"><Spinner />{text('loading')}</div> : !filtered.length ? <div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text(items.length ? 'noMatchingBadges' : 'empty')}</div> : <div className="divide-y">
        {filtered.map(item => <article key={item.code} className="flex flex-wrap items-center gap-3 p-3 sm:gap-4" aria-label={item.name}>
          <ForumBadge badge={item} size="md" className={!item.isEnabled ? 'opacity-50' : undefined} />
          <div className="min-w-0 flex-[1_1_12rem]"><button type="button" onClick={() => setEditing(item)} className="max-w-full break-words text-left text-sm font-semibold hover:underline">{item.name}</button><p className="mt-0.5 line-clamp-2 break-words text-xs text-muted-foreground">{item.description || text('noBadgeDescription')}</p><div className="mt-2 flex flex-wrap gap-x-2 gap-y-1 text-xs text-muted-foreground"><span>{text(item.type === 'system' ? 'system' : 'custom')}</span><span aria-hidden>·</span><span>{badgeLevelLabel(item.level, text)}</span><span aria-hidden>·</span><span>{text(item.grantMode === 'auto' ? 'auto' : 'manual')}</span><Badge variant={item.isEnabled ? 'secondary' : 'outline'} className="text-[10px]">{text(item.isEnabled ? 'enabled' : 'disabled')}</Badge></div></div>
          <div className="flex w-full items-center gap-1 border-t pt-2 sm:w-auto sm:border-t-0 sm:pt-0"><Button variant="ghost" size="sm" className="mr-auto sm:mr-2" disabled={!canManageRecipients} title={!canManageRecipients ? text('requiresUserManagement') : undefined} onClick={() => setMembers(item)}><Users data-icon="inline-start" />{text('recipients')} {item.grantedCount ?? 0}</Button><Button variant="ghost" size="icon-sm" aria-label={`${text('edit')} ${item.name}`} onClick={() => setEditing(item)}><Edit3 /></Button>{item.type !== 'system' ? <span title={item.canDelete === false ? text('badgeDeleteBlocked') : text('delete')}><Button variant="ghost" size="icon-sm" disabled={item.canDelete === false} aria-label={`${text('delete')} ${item.name}`} onClick={() => setDeleting(item)}><Trash2 /></Button></span> : null}</div>
        </article>)}
      </div>}
    </section>
    {editing ? <BadgeEditor value={editing} api={api} text={text} onClose={() => setEditing(null)} onSaved={load} /> : null}
    {members && canManageRecipients ? <BadgeRecipients badge={items.find(item => item.code === members.code) || members} api={api} text={text} onClose={() => setMembers(null)} onChanged={load} /> : null}
    {deleting ? <DeleteBadge value={deleting} api={api} text={text} onClose={() => setDeleting(null)} onDeleted={load} /> : null}
  </AdminPage>
}

function DeleteBadge({ value, api, text, onClose, onDeleted }: { value: AdminBadge; api: GooseAdminApi; text: BadgeText; onClose(): void; onDeleted(): Promise<void> }) {
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  async function remove() {
    setSaving(true)
    setError('')
    try { await api.assets.deleteBadge(value.code); toast.success(text('saved')); await onDeleted(); onClose() }
    catch (reason) { setError(errorMessage(reason, text('saveFailed'))) }
    finally { setSaving(false) }
  }
  return <Dialog open onOpenChange={open => { if (!open && !saving) onClose() }}><DialogContent showCloseButton={!saving}><DialogHeader><DialogTitle>{text('deleteTitle')}</DialogTitle><DialogDescription>{value.name} · {text('badgeDeleteHint')}</DialogDescription></DialogHeader>{error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}<DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button variant="destructive" disabled={saving} onClick={() => void remove()}>{saving ? <Spinner data-icon="inline-start" /> : null}{text('delete')}</Button></DialogFooter></DialogContent></Dialog>
}
