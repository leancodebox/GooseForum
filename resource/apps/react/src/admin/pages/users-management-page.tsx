import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminBadge, AdminUser, GooseAdminApi, UserBadge } from '@gooseforum/client'
import { Alert, AlertDescription, AlertTitle } from '@gooseforum/react/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@gooseforum/react/components/ui/avatar'
import { Badge } from '@gooseforum/react/components/ui/badge'
import { Button } from '@gooseforum/react/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/react/components/ui/dialog'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/react/components/ui/empty'
import { Field, FieldGroup, FieldLabel } from '@gooseforum/react/components/ui/field'
import { Input } from '@gooseforum/react/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/react/components/ui/select'
import { Spinner } from '@gooseforum/react/components/ui/spinner'
import { Switch } from '@gooseforum/react/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/react/components/ui/table'
import { AlertTriangle, Award, CheckCircle2, ChevronLeft, ChevronRight, RefreshCw, Search, ShieldOff, UserCog } from 'lucide-react'
import { toast } from 'sonner'
import type { UserTextKey } from '../users-i18n'

type Text = (key: UserTextKey) => string

export function UsersManagementPage({ api, text }: { api: GooseAdminApi; text: Text }) {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [search, setSearch] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null)

  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const rangeStart = total ? (page - 1) * pageSize + 1 : 0
  const rangeEnd = Math.min(page * pageSize, total)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const result = await api.users.list({ page, pageSize, username: appliedSearch || undefined })
      setUsers(result.list || [])
      setTotal(result.total || 0)
    } catch (reason) {
      setError(errorMessage(reason, text('loadFailed')))
    } finally {
      setLoading(false)
    }
  }, [api, appliedSearch, page, pageSize, text])

  useEffect(() => { void load() }, [load])
  useEffect(() => { if (page > totalPages) setPage(totalPages) }, [page, totalPages])

  function applySearch(event?: React.FormEvent) {
    event?.preventDefault()
    setAppliedSearch(search.trim())
    setPage(1)
  }

  return <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
    <header><h2 className="text-lg font-semibold tracking-tight">{text('title')}</h2><p className="text-xs text-muted-foreground">{text('description')}</p></header>
    {error ? <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('loadFailed')}</AlertTitle><AlertDescription className="flex items-center justify-between gap-3"><span>{error}</span><Button variant="outline" size="sm" onClick={() => void load()}>{text('retry')}</Button></AlertDescription></Alert> : null}
    <section className="overflow-hidden rounded-lg border bg-background">
      <div className="flex flex-col gap-2 border-b bg-muted/20 p-2 lg:flex-row lg:items-center lg:justify-between">
        <form className="flex min-w-0 flex-1 items-center gap-1.5 lg:max-w-md" onSubmit={applySearch}><div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} className="h-8 pl-8" placeholder={text('search')} /></div><Button size="sm" type="submit">{text('searchAction')}</Button>{appliedSearch ? <Button variant="ghost" size="sm" type="button" onClick={() => { setSearch(''); setAppliedSearch(''); setPage(1) }}>{text('clear')}</Button> : null}</form>
        <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground"><Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button><span className="whitespace-nowrap">{rangeStart}-{rangeEnd} / {total}</span><Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1) }}><SelectTrigger size="sm" className="w-20"><SelectValue /></SelectTrigger><SelectContent><SelectGroup>{[10, 20, 30, 50].map((size) => <SelectItem key={size} value={String(size)}>{size}</SelectItem>)}</SelectGroup></SelectContent></Select><Button variant="outline" size="icon-sm" disabled={page <= 1 || loading} title={text('previous')} onClick={() => setPage(page - 1)}><ChevronLeft /></Button><span className="min-w-12 text-center">{page}/{totalPages}</span><Button variant="outline" size="icon-sm" disabled={page >= totalPages || loading} title={text('next')} onClick={() => setPage(page + 1)}><ChevronRight /></Button></div>
      </div>
      {loading && !users.length ? <UserEmpty icon={<Spinner />} title={text('loading')} /> : !users.length ? <UserEmpty icon={<Search />} title={text('empty')} /> : <><div className="hidden md:block"><Table className="min-w-220 table-fixed"><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8">{text('user')}</TableHead><TableHead className="h-8 w-48">{text('roles')}</TableHead><TableHead className="h-8 w-28">{text('status')}</TableHead><TableHead className="h-8 w-40">{text('createdAt')}</TableHead><TableHead className="h-8 w-40">{text('lastActive')}</TableHead><TableHead className="h-8 w-16 text-right">{text('actions')}</TableHead></TableRow></TableHeader><TableBody>{users.map((user) => <TableRow key={user.userId}><TableCell className="py-2"><UserIdentity user={user} text={text} /></TableCell><TableCell className="py-2"><div className="flex flex-wrap gap-1">{user.roleList?.length ? user.roleList.map((role) => <Badge key={role.value} variant="secondary">{role.name}</Badge>) : <span className="text-xs text-muted-foreground">{text('noRole')}</span>}</div></TableCell><TableCell className="py-2"><Badge variant={user.status === 0 ? 'outline' : 'destructive'}>{user.status === 0 ? <CheckCircle2 data-icon="inline-start" /> : <ShieldOff data-icon="inline-start" />}{user.status === 0 ? text('enabled') : text('disabled')}</Badge></TableCell><TableCell className="truncate py-2 text-xs text-muted-foreground">{user.createTime || '—'}</TableCell><TableCell className="truncate py-2 text-xs text-muted-foreground">{user.lastActiveTime || text('never')}</TableCell><TableCell className="py-2 text-right"><Button variant="ghost" size="icon-sm" title={text('edit')} onClick={() => setEditingUser(user)}><UserCog /></Button></TableCell></TableRow>)}</TableBody></Table></div><div className="divide-y md:hidden">{users.map((user) => <article key={user.userId} className="flex items-start gap-2.5 px-3 py-2.5"><UserIdentity user={user} text={text} compact /><Button variant="ghost" size="icon-sm" title={text('edit')} onClick={() => setEditingUser(user)}><UserCog /></Button></article>)}</div></>}
    </section>
    <UserEditor user={editingUser} api={api} text={text} onClose={() => setEditingUser(null)} onSaved={async () => { setEditingUser(null); await load() }} />
  </main>
}

function UserIdentity({ user, text, compact = false }: { user: AdminUser; text: Text; compact?: boolean }) {
  return <div className="flex min-w-0 flex-1 items-center gap-2.5"><a href={`/u/${user.userId}`} target="_blank" rel="noreferrer" className="shrink-0"><Avatar className={compact ? 'size-10' : 'size-9'}><AvatarImage src={user.avatarUrl || undefined} alt={user.username} /><AvatarFallback>{user.username.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar></a><div className="min-w-0 flex-1"><a href={`/u/${user.userId}`} target="_blank" rel="noreferrer" className="block truncate font-semibold hover:text-primary hover:underline">{user.username}</a><div className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground"><span className="truncate">{user.email || '—'}</span><Badge variant="outline" className="shrink-0">{user.validate === 1 ? text('verified') : text('unverified')}</Badge></div>{compact ? <div className="mt-1 flex flex-wrap gap-1"><Badge variant={user.status === 0 ? 'outline' : 'destructive'}>{user.status === 0 ? text('enabled') : text('disabled')}</Badge>{user.roleList?.map((role) => <Badge key={role.value} variant="secondary">{role.name}</Badge>)}</div> : null}</div></div>
}

function UserEditor({ user, api, text, onClose, onSaved }: { user: AdminUser | null; api: GooseAdminApi; text: Text; onClose(): void; onSaved(): Promise<void> }) {
  const [form, setForm] = useState({ status: 0, validate: 0, roleId: 0 })
  const [roles, setRoles] = useState<{ name: string; value: number }[]>([])
  const [badgeOptions, setBadgeOptions] = useState<AdminBadge[]>([])
  const [activeBadges, setActiveBadges] = useState<UserBadge[]>([])
  const [selectedBadgeCodes, setSelectedBadgeCodes] = useState<string[]>([])
  const [badgeLoading, setBadgeLoading] = useState(false)
  const [badgesLoaded, setBadgesLoaded] = useState(false)
  const [saving, setSaving] = useState(false)
  const textRef = useRef(text)
  useEffect(() => { textRef.current = text }, [text])

  useEffect(() => {
    if (!user) return
    let active = true
    setForm({ status: user.status, validate: user.validate, roleId: user.roleId || 0 })
    setBadgeLoading(true); setBadgesLoaded(false); setBadgeOptions([]); setActiveBadges([]); setSelectedBadgeCodes([])
    void api.users.roles().then((roleItems) => {
      if (active) setRoles(roleItems)
    }).catch(() => {
      if (active) setRoles([])
    })
    void api.users.badgeOptions(user.userId).then((badges) => {
      if (!active) return
      setBadgeOptions(badges.options || [])
      setActiveBadges(badges.active || [])
      const optionCodes = new Set((badges.options || []).map((badge) => badge.code))
      setSelectedBadgeCodes((badges.active || []).filter((badge) => badge.source === 'manual' && optionCodes.has(badge.code)).map((badge) => badge.code))
      setBadgesLoaded(true)
    }).catch((reason) => { if (active) toast.error(errorMessage(reason, textRef.current('badgeLoading'))) }).finally(() => { if (active) setBadgeLoading(false) })
    return () => { active = false }
  }, [api, user])

  const automaticBadges = activeBadges.filter((badge) => badge.source !== 'manual')
  function toggleBadge(code: string) { setSelectedBadgeCodes((current) => current.includes(code) ? current.filter((item) => item !== code) : [...current, code]) }
  async function save(event: React.FormEvent) {
    event.preventDefault()
    if (!user) return
    setSaving(true)
    try { await api.users.edit({ userId: user.userId, ...form }); if (badgesLoaded) await api.users.saveBadges(user.userId, selectedBadgeCodes); toast.success(text('saved')); await onSaved() } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }

  return <Dialog open={Boolean(user)} onOpenChange={(open) => !open && !saving && onClose()}><DialogContent className="flex max-h-[min(700px,calc(100vh-1.5rem))] gap-0 overflow-hidden p-0 sm:max-w-4xl"><form onSubmit={save} className="flex min-h-0 w-full flex-col"><DialogHeader className="border-b px-4 py-3"><div className="flex items-center gap-3">{user ? <Avatar className="size-10"><AvatarImage src={user.avatarUrl || undefined} alt={user.username} /><AvatarFallback>{user.username.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar> : null}<div className="min-w-0"><DialogTitle>{text('edit')}</DialogTitle><DialogDescription className="truncate">{user?.username} · {user?.email || '—'}</DialogDescription></div></div></DialogHeader><div className="min-h-0 flex-1 overflow-y-auto p-3"><div className="grid gap-3 lg:grid-cols-[280px_minmax(0,1fr)]"><aside className="flex flex-col gap-3"><section className="rounded-lg border p-3"><h3 className="mb-2 text-sm font-semibold">{text('account')}</h3><FieldGroup><Field orientation="horizontal"><FieldLabel>{text('status')}</FieldLabel><span className="ml-auto flex items-center gap-2 text-xs text-muted-foreground">{form.status === 0 ? text('enabled') : text('disabled')}<Switch checked={form.status === 0} onCheckedChange={(checked) => setForm({ ...form, status: checked ? 0 : 1 })} /></span></Field><Field orientation="horizontal"><FieldLabel>{text('emailVerification')}</FieldLabel><span className="ml-auto flex items-center gap-2 text-xs text-muted-foreground">{form.validate === 1 ? text('verified') : text('unverified')}<Switch checked={form.validate === 1} onCheckedChange={(checked) => setForm({ ...form, validate: checked ? 1 : 0 })} /></span></Field><Field><FieldLabel>{text('role')}</FieldLabel><Select value={String(form.roleId)} onValueChange={(value) => setForm({ ...form, roleId: Number(value) })}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="0">{text('noRole')}</SelectItem>{roles.map((role) => <SelectItem key={role.value} value={String(role.value)}>{role.name}</SelectItem>)}</SelectGroup></SelectContent></Select></Field></FieldGroup></section><section className="rounded-lg border p-3"><h3 className="mb-2 text-sm font-semibold">{text('details')}</h3><dl className="grid gap-2 text-sm"><Detail label={text('createdAt')} value={user?.createTime || '—'} /><Detail label={text('lastActive')} value={user?.lastActiveTime || text('never')} /><Detail label={text('prestige')} value={String(user?.prestige || 0)} /></dl></section></aside><section className="rounded-lg border"><div className="flex items-center justify-between gap-3 border-b px-3 py-2.5"><div><h3 className="flex items-center gap-2 text-sm font-semibold"><Award className="size-4 text-muted-foreground" />{text('badges')}</h3><p className="text-xs text-muted-foreground">{text('badgeHint')}</p></div>{badgeLoading ? <Spinner /> : null}</div><div className="flex flex-col gap-3 p-3">{automaticBadges.length ? <div className="flex flex-col gap-2"><h4 className="text-xs font-medium text-muted-foreground">{text('automaticBadges')}</h4><div className="flex flex-wrap gap-1.5">{automaticBadges.map((badge) => <Badge key={badge.code} variant="secondary">{badge.name}</Badge>)}</div></div> : null}<div className="flex flex-col gap-2"><h4 className="text-xs font-medium text-muted-foreground">{text('manualBadges')}</h4>{badgeOptions.length ? <div className="grid grid-cols-3 gap-2 sm:grid-cols-4 xl:grid-cols-6">{badgeOptions.map((badge) => <Button key={badge.code} type="button" variant="ghost" data-current={selectedBadgeCodes.includes(badge.code)} title={badge.description || badge.code} onClick={() => toggleBadge(badge.code)} className="relative h-auto min-h-20 min-w-0 flex-col gap-1 whitespace-normal rounded-md border border-transparent bg-muted/40 p-2 font-normal data-[current=true]:border-primary data-[current=true]:bg-primary/5"><span className="grid size-9 place-items-center rounded-full bg-background ring-1 ring-border"><img src={badge.iconUrl || '/static/badges/contributor.svg'} alt="" className="size-5 object-contain" /></span><span className="max-w-full truncate text-xs font-medium">{badge.name}</span>{selectedBadgeCodes.includes(badge.code) ? <span className="absolute right-1 top-1 text-primary">✓</span> : null}</Button>)}</div> : <Empty className="min-h-28 border-dashed p-3"><EmptyHeader><EmptyTitle>{badgeLoading ? text('badgeLoading') : text('noBadges')}</EmptyTitle></EmptyHeader></Empty>}</div></div></section></div></div><DialogFooter><Button variant="outline" type="button" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button type="submit" disabled={saving}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button></DialogFooter></form></DialogContent></Dialog>
}

function Detail({ label, value }: { label: string; value: string }) { return <div className="flex justify-between gap-3"><dt className="text-muted-foreground">{label}</dt><dd className="truncate text-right">{value}</dd></div> }
function UserEmpty({ icon, title }: { icon: React.ReactNode; title: string }) { return <Empty className="min-h-32 rounded-none border-0"><EmptyHeader><EmptyMedia variant="icon">{icon}</EmptyMedia><EmptyTitle>{title}</EmptyTitle></EmptyHeader></Empty> }
function errorMessage(reason: unknown, fallback: string) { return reason instanceof Error && reason.message ? reason.message : fallback }
