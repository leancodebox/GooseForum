import { AdminPage } from '../components/admin-page'
import { UserMFAReset } from '../components/user-mfa-reset'
import { RestrictionFields, RestrictionHistory } from '../components/user-restriction'
import { type RestrictionForm, restrictionForm } from '../components/user-restriction-form'
import { useLatestRequest } from '../use-latest-request'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminBadge, AdminUser, GooseAdminApi, UserBadge } from '@gooseforum/client'
import { Alert, AlertDescription, AlertTitle } from '@gooseforum/ui/components/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@gooseforum/ui/components/avatar'
import { Badge } from '@gooseforum/ui/components/badge'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/ui/components/empty'
import { Field, FieldGroup, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Switch } from '@gooseforum/ui/components/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@gooseforum/ui/components/tabs'
import { AlertTriangle, CheckCircle2, ChevronLeft, ChevronRight, RefreshCw, Search, ShieldOff, UserCog } from 'lucide-react'
import { toast } from 'sonner'
import type { UserTextKey } from '../users-i18n'

type Text = (key: UserTextKey) => string

export function UsersManagementPage({ api, text, locale = 'en', canResetMFA = false, currentUserId }: { api: GooseAdminApi; text: Text; locale?: string; canResetMFA?: boolean; currentUserId?: number }) {
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

  const beginRequest = useLatestRequest();
  const load = useCallback(async () => {
    const isCurrent = beginRequest();
    setLoading(true)
    setError('')
    try {
      const result = await api.users.list({ page, pageSize, username: appliedSearch || undefined })
      if (!isCurrent()) return;
      setUsers(result.list || [])
      setTotal(result.total || 0)
    } catch (reason) {
      if (!isCurrent()) return;
      setError(errorMessage(reason, text('loadFailed')))
    } finally {
      if (isCurrent()) setLoading(false)
    }
  }, [beginRequest, api, appliedSearch, page, pageSize, text])

  useEffect(() => { void load() }, [load])
  useEffect(() => { if (page > totalPages) setPage(totalPages) }, [page, totalPages])

  function applySearch(event?: React.FormEvent) {
    event?.preventDefault()
    setAppliedSearch(search.trim())
    setPage(1)
  }

  return <AdminPage>
    <header><h2 className="text-lg font-semibold tracking-tight">{text('title')}</h2><p className="text-xs text-muted-foreground">{text('description')}</p></header>
    {error ? <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('loadFailed')}</AlertTitle><AlertDescription className="flex items-center justify-between gap-3"><span>{error}</span><Button variant="outline" size="sm" onClick={() => void load()}>{text('retry')}</Button></AlertDescription></Alert> : null}
    <section className="overflow-hidden rounded-lg border bg-background">
      <div className="flex flex-col gap-2 border-b bg-muted/20 p-2 lg:flex-row lg:items-center lg:justify-between">
        <form className="flex min-w-0 flex-1 items-center gap-1.5 lg:max-w-md" onSubmit={applySearch}><div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} className="h-8 pl-8" placeholder={text('search')} /></div><Button size="sm" type="submit">{text('searchAction')}</Button>{appliedSearch ? <Button variant="ghost" size="sm" type="button" onClick={() => { setSearch(''); setAppliedSearch(''); setPage(1) }}>{text('clear')}</Button> : null}</form>
        <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground"><Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button><span className="whitespace-nowrap">{rangeStart}-{rangeEnd} / {total}</span><Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1) }}><SelectTrigger size="sm" className="w-20"><SelectValue /></SelectTrigger><SelectContent><SelectGroup>{[10, 20, 30, 50].map((size) => <SelectItem key={size} value={String(size)}>{size}</SelectItem>)}</SelectGroup></SelectContent></Select><Button variant="outline" size="icon-sm" disabled={page <= 1 || loading} title={text('previous')} onClick={() => setPage(page - 1)}><ChevronLeft /></Button><span className="min-w-12 text-center">{page}/{totalPages}</span><Button variant="outline" size="icon-sm" disabled={page >= totalPages || loading} title={text('next')} onClick={() => setPage(page + 1)}><ChevronRight /></Button></div>
      </div>
      {loading && !users.length ? <UserEmpty icon={<Spinner />} title={text('loading')} /> : !users.length ? <UserEmpty icon={<Search />} title={text('empty')} /> : <><div className="hidden md:block"><Table className="min-w-220 table-fixed"><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8">{text('user')}</TableHead><TableHead className="h-8 w-48">{text('roles')}</TableHead><TableHead className="h-8 w-28">{text('status')}</TableHead><TableHead className="h-8 w-40">{text('createdAt')}</TableHead><TableHead className="h-8 w-40">{text('lastActive')}</TableHead><TableHead className="h-8 w-16 text-right">{text('actions')}</TableHead></TableRow></TableHeader><TableBody>{users.map((user) => <TableRow key={user.userId}><TableCell className="py-2"><UserIdentity user={user} text={text} /></TableCell><TableCell className="py-2"><div className="flex flex-wrap gap-1">{user.roleList?.length ? user.roleList.map((role) => <Badge key={role.value} variant="secondary">{role.name}</Badge>) : <span className="text-xs text-muted-foreground">{text('noRole')}</span>}</div></TableCell><TableCell className="py-2"><Badge variant={(user.restrictionStatus || 'normal') === 'normal' ? 'outline' : 'destructive'}>{(user.restrictionStatus || 'normal') === 'normal' ? <CheckCircle2 data-icon="inline-start" /> : <ShieldOff data-icon="inline-start" />}{(user.restrictionStatus || 'normal') === 'normal' ? text('enabled') : user.restrictionStatus === 'banned' ? text('banned') : text('suspended')}</Badge></TableCell><TableCell className="truncate py-2 text-xs text-muted-foreground">{user.createTime || '—'}</TableCell><TableCell className="truncate py-2 text-xs text-muted-foreground">{user.lastActiveTime || text('never')}</TableCell><TableCell className="py-2 text-right"><Button variant="ghost" size="icon-sm" title={text('edit')} onClick={() => setEditingUser(user)}><UserCog /></Button></TableCell></TableRow>)}</TableBody></Table></div><div className="divide-y md:hidden">{users.map((user) => <article key={user.userId} className="flex items-start gap-2.5 px-3 py-2.5"><UserIdentity user={user} text={text} compact /><Button variant="ghost" size="icon-sm" title={text('edit')} onClick={() => setEditingUser(user)}><UserCog /></Button></article>)}</div></>}
    </section>
    <UserEditor key={editingUser?.userId || 'closed'} user={editingUser} api={api} text={text} locale={locale} canResetMFA={canResetMFA && editingUser?.userId !== currentUserId} onClose={() => setEditingUser(null)} onSaved={async () => { setEditingUser(null); await load() }} />
  </AdminPage>
}

function UserIdentity({ user, text, compact = false }: { user: AdminUser; text: Text; compact?: boolean }) {
  return <div className="flex min-w-0 flex-1 items-center gap-2.5"><a href={`/u/${user.userId}`} target="_blank" rel="noreferrer" className="shrink-0"><Avatar className={compact ? 'size-10' : 'size-9'}><AvatarImage src={user.avatarUrl || undefined} alt={user.username} /><AvatarFallback>{user.username.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar></a><div className="min-w-0 flex-1"><a href={`/u/${user.userId}`} target="_blank" rel="noreferrer" className="block truncate font-semibold hover:text-primary hover:underline">{user.username}</a><div className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground"><span className="truncate">{user.email || '—'}</span><Badge variant="outline" className="shrink-0">{user.validate === 1 ? text('verified') : text('unverified')}</Badge></div>{compact ? <div className="mt-1 flex flex-wrap gap-1"><Badge variant={(user.restrictionStatus || 'normal') === 'normal' ? 'outline' : 'destructive'}>{(user.restrictionStatus || 'normal') === 'normal' ? text('enabled') : user.restrictionStatus === 'banned' ? text('banned') : text('suspended')}</Badge>{user.roleList?.map((role) => <Badge key={role.value} variant="secondary">{role.name}</Badge>)}</div> : null}</div></div>
}

function UserEditor({ user, api, text, locale, canResetMFA, onClose, onSaved }: { user: AdminUser | null; api: GooseAdminApi; text: Text; locale: string; canResetMFA: boolean; onClose(): void; onSaved(): Promise<void> }) {
  const [form, setForm] = useState(() => ({ validate: user?.validate || 0, roleId: user?.roleId || 0 }))
  const [restriction, setRestriction] = useState<RestrictionForm>(() => restrictionForm(user))
  const [roles, setRoles] = useState<{ name: string; value: number }[]>([])
  const [rolesLoaded, setRolesLoaded] = useState(false)
  const [roleLoading, setRoleLoading] = useState(true)
  const [roleError, setRoleError] = useState('')
  const [roleRetry, setRoleRetry] = useState(0)
  const [badgeOptions, setBadgeOptions] = useState<AdminBadge[]>([])
  const [activeBadges, setActiveBadges] = useState<UserBadge[]>([])
  const [selectedBadgeCodes, setSelectedBadgeCodes] = useState<string[]>([])
  const [initialBadgeCodes, setInitialBadgeCodes] = useState<string[]>([])
  const [badgeLoading, setBadgeLoading] = useState(true)
  const [badgesLoaded, setBadgesLoaded] = useState(false)
  const [badgeError, setBadgeError] = useState('')
  const [badgeRetry, setBadgeRetry] = useState(0)
  const [saving, setSaving] = useState(false)
  const [editorTab, setEditorTab] = useState('account')
  const [validationError, setValidationError] = useState('')
  const textRef = useRef(text)
  useEffect(() => { textRef.current = text }, [text])

  useEffect(() => {
    if (!user) return
    let active = true
    void Promise.resolve().then(() => {
      if (!active) return
      setRoleLoading(true); setRolesLoaded(false); setRoleError(''); setRoles([])
      return api.users.roles()
    }).then((roleItems) => {
      if (!active || !roleItems) return
      setRoles(roleItems)
      setRolesLoaded(true)
    }).catch((reason) => { if (active) setRoleError(errorMessage(reason, textRef.current('roleLoadFailed'))) }).finally(() => { if (active) setRoleLoading(false) })
    return () => { active = false }
  }, [api, user, roleRetry])

  useEffect(() => {
    if (!user) return
    let active = true
    void Promise.resolve().then(() => {
      if (!active) return
      setBadgeLoading(true); setBadgesLoaded(false); setBadgeError(''); setBadgeOptions([]); setActiveBadges([]); setSelectedBadgeCodes([])
      return api.users.badgeOptions(user.userId)
    }).then((badges) => {
      if (!active || !badges) return
      setBadgeOptions(badges.options || [])
      setActiveBadges(badges.active || [])
      const optionCodes = new Set((badges.options || []).map((badge) => badge.code))
      const codes = (badges.active || []).filter((badge) => badge.source === 'manual' && optionCodes.has(badge.code)).map((badge) => badge.code)
      setSelectedBadgeCodes(codes)
      setInitialBadgeCodes(codes)
      setBadgesLoaded(true)
    }).catch((reason) => { if (active) setBadgeError(errorMessage(reason, textRef.current('badgeLoadFailed'))) }).finally(() => { if (active) setBadgeLoading(false) })
    return () => { active = false }
  }, [api, user, badgeRetry])

  const roleName = roles.find((role) => role.value === form.roleId)?.name || user?.roleList?.find((role) => role.value === form.roleId)?.name || (form.roleId === 0 ? text('noRole') : `#${form.roleId}`)
  const automaticBadges = activeBadges.filter((badge) => badge.source !== 'manual')
  function toggleBadge(code: string) { setSelectedBadgeCodes((current) => current.includes(code) ? current.filter((item) => item !== code) : [...current, code]) }
  async function save(event: React.FormEvent) {
    event.preventDefault()
    if (!user || saving) return
    if ((restriction.status !== 'normal' || restrictionForm(user).status !== 'normal') && !restriction.reason.trim()) {
      setEditorTab('account')
      setValidationError(text('reasonRequired'))
      return
    }
    setSaving(true)
    try {
      await api.users.edit({ userId: user.userId, ...form, restrictionStatus: restriction.status, restrictionUntil: restriction.until && restriction.status !== 'normal' ? new Date(restriction.until).toISOString() : null, restrictionReason: restriction.reason })
      const badgesChanged = selectedBadgeCodes.length !== initialBadgeCodes.length || selectedBadgeCodes.some((code) => !initialBadgeCodes.includes(code))
      if (badgesLoaded && badgesChanged) {
        await api.users.saveBadges(user.userId, selectedBadgeCodes)
        setInitialBadgeCodes(selectedBadgeCodes)
      }
      toast.success(text('saved'))
      await onSaved()
    } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }

  return <Dialog open={Boolean(user)} onOpenChange={(open) => !open && !saving && onClose()}>
    <DialogContent className="flex max-h-[calc(100dvh-2rem)] gap-0 overflow-hidden p-0 sm:max-w-2xl">
      <form noValidate onSubmit={save} className="flex min-h-0 w-full flex-col">
        <DialogHeader className="border-b px-4 py-3">
          <div className="flex items-center gap-3">
            {user ? <Avatar className="size-10"><AvatarImage src={user.avatarUrl || undefined} alt={user.username} /><AvatarFallback>{user.username.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar> : null}
            <div className="min-w-0 pr-6">
              <DialogTitle>{text('edit')}</DialogTitle>
              <DialogDescription className="truncate">{user?.username} · {user?.email || '—'}</DialogDescription>
            </div>
          </div>
        </DialogHeader>
        <Tabs value={editorTab} onValueChange={setEditorTab} className="min-h-0 flex-1 gap-0">
          <TabsList variant="line" className="mx-4 mt-2 grid w-auto shrink-0 grid-cols-3">
            <TabsTrigger type="button" value="account">{text('account')}</TabsTrigger>
            <TabsTrigger type="button" value="badges">{text('badges')}</TabsTrigger>
            <TabsTrigger type="button" value="history">{text('historyTab')}</TabsTrigger>
          </TabsList>
          <div className="min-h-0 flex-1 overflow-y-auto p-4 sm:min-h-[min(20rem,calc(100dvh-13rem))]">
            <TabsContent value="account">
              <FieldGroup className="grid items-start gap-4 sm:grid-cols-3 [&>[data-slot=field]:has(#restriction-status)]:order-2 [&>[data-slot=field]:has(#restriction-until)]:order-4">
                <Field className="order-1">
                  <FieldLabel htmlFor="user-role">{text('role')}</FieldLabel>
                  <Select disabled={!rolesLoaded || saving} value={String(form.roleId)} onValueChange={(value) => setForm({ ...form, roleId: Number(value) })}>
                    <SelectTrigger id="user-role" className="w-full"><SelectValue>{roleName}</SelectValue></SelectTrigger>
                    <SelectContent><SelectGroup><SelectItem value="0">{text('noRole')}</SelectItem>{roles.map((role) => <SelectItem key={role.value} value={String(role.value)}>{role.name}</SelectItem>)}</SelectGroup></SelectContent>
                  </Select>
                  {roleLoading ? <p className="text-xs text-muted-foreground">{text('loading')}</p> : roleError ? <div role="alert" className="flex items-start gap-2"><p className="min-w-0 flex-1 break-words text-xs">{roleError}</p><Button type="button" variant="outline" size="sm" disabled={saving} onClick={() => setRoleRetry((value) => value + 1)}><RefreshCw data-icon="inline-start" />{text('retry')}</Button></div> : null}
                </Field>
                <Field className="order-3">
                  <FieldLabel htmlFor="user-email-verification">{text('emailVerification')}</FieldLabel>
                  <span className="flex h-9 items-center justify-between gap-2 text-xs text-muted-foreground">
                    {form.validate === 1 ? text('verified') : text('unverified')}
                    <Switch id="user-email-verification" checked={form.validate === 1} onCheckedChange={(checked) => setForm({ ...form, validate: checked ? 1 : 0 })} />
                  </span>
                </Field>
                <RestrictionFields value={restriction} initialStatus={restrictionForm(user).status} onChange={(value) => { setRestriction(value.status !== restriction.status ? { ...value, reason: value.status === restrictionForm(user).status ? restrictionForm(user).reason : '' } : value); setValidationError('') }} text={text} locale={locale} reasonClassName="order-5 sm:col-span-3" />
                {validationError ? <p role="alert" className="order-6 text-sm sm:col-span-3">{validationError}</p> : null}
              </FieldGroup>
              {user && canResetMFA ? <UserMFAReset userId={user.userId} username={user.username} api={api} text={text} disabled={saving} /> : null}
              <dl className="mt-4 grid grid-cols-3 gap-3 border-t pt-3 text-xs">
                <Detail label={text('createdAt')} value={user?.createTime || '—'} />
                <Detail label={text('lastActive')} value={user?.lastActiveTime || text('never')} />
                <Detail label={text('prestige')} value={String(user?.prestige || 0)} />
              </dl>
            </TabsContent>
            <TabsContent value="badges">
              <div className="flex flex-col gap-3">
                {automaticBadges.length ? <div className="flex flex-col gap-2"><h4 className="text-xs font-medium text-muted-foreground">{text('automaticBadges')}</h4><div className="flex flex-wrap gap-1.5">{automaticBadges.map((badge) => <Badge key={badge.code} variant="secondary">{badge.name}</Badge>)}</div></div> : null}
                <div className="flex flex-col gap-2">
                  <h4 className="flex items-center justify-between text-xs font-medium text-muted-foreground">{text('manualBadges')}{badgeLoading ? <Spinner /> : null}</h4>
                  {badgeError ? <div role="alert" className="flex items-start gap-2"><p className="min-w-0 flex-1 break-words text-xs">{badgeError}</p><Button type="button" variant="outline" size="sm" disabled={saving} onClick={() => setBadgeRetry((value) => value + 1)}><RefreshCw data-icon="inline-start" />{text('retry')}</Button></div> : badgeOptions.length ? <div className="grid grid-cols-[repeat(auto-fill,minmax(4.5rem,1fr))] gap-1.5">
                    {badgeOptions.map((badge) => <Button key={badge.code} type="button" variant="ghost" disabled={saving} aria-pressed={selectedBadgeCodes.includes(badge.code)} data-current={selectedBadgeCodes.includes(badge.code)} title={badge.description ? `${badge.name}: ${badge.description}` : badge.name} onClick={() => toggleBadge(badge.code)} className="relative h-auto min-h-18 min-w-0 flex-col gap-1 whitespace-normal rounded-md border border-transparent bg-muted/40 p-1.5 font-normal data-[current=true]:border-primary data-[current=true]:bg-primary/5">
                      <span className="grid size-9 place-items-center rounded-full bg-background ring-1 ring-border"><img src={badge.iconUrl || '/static/badges/contributor.svg'} alt="" className="size-5 object-contain" /></span>
                      <span className="max-w-full truncate text-xs font-medium">{badge.name}</span>
                      {selectedBadgeCodes.includes(badge.code) ? <CheckCircle2 aria-hidden="true" className="absolute right-1 top-1 size-3.5 text-primary" /> : null}
                    </Button>)}
                  </div> : <p className="text-xs text-muted-foreground">{badgeLoading ? text('badgeLoading') : text('noBadges')}</p>}
                </div>
              </div>
            </TabsContent>
            <TabsContent value="history">
              <RestrictionHistory userId={user?.userId} api={api} text={text} className="mt-0 min-w-0 border-t-0 pt-0" />
            </TabsContent>
          </div>
        </Tabs>
        <DialogFooter className="mx-0 mb-0 rounded-b-none">
          <Button variant="outline" type="button" disabled={saving} onClick={onClose}>{text('cancel')}</Button>
          <Button type="submit" disabled={saving}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
}

function Detail({ label, value }: { label: string; value: string }) { return <div className="min-w-0"><dt className="text-muted-foreground">{label}</dt><dd className="mt-1 truncate" title={value}>{value}</dd></div> }
function UserEmpty({ icon, title }: { icon: React.ReactNode; title: string }) { return <Empty className="min-h-32 rounded-none border-0"><EmptyHeader><EmptyMedia variant="icon">{icon}</EmptyMedia><EmptyTitle>{title}</EmptyTitle></EmptyHeader></Empty> }
function errorMessage(reason: unknown, fallback: string) { return reason instanceof Error && reason.message ? reason.message : fallback }
