import { ColorPicker } from '@gooseforum/ui/components/color-picker'
import { useLatestRequest } from '../use-latest-request'
import { useCallback, useEffect, useMemo, useState } from 'react'
import type { AccessControlOverview, AdminCategory, AdminCategoryModerator, AdminUser, GooseAdminApi } from '@gooseforum/client'
import { Alert, AlertDescription, AlertTitle } from '@gooseforum/ui/components/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@gooseforum/ui/components/avatar'
import { Badge } from '@gooseforum/ui/components/badge'
import { Button } from '@gooseforum/ui/components/button'
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@gooseforum/ui/components/field'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/ui/components/empty'
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuTrigger } from '@gooseforum/ui/components/dropdown-menu'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@gooseforum/ui/components/sheet'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table'
import { AlertTriangle, LoaderCircle, LockKeyhole, MoreHorizontal, Pencil, Plus, Search, Shield, Trash2, UserPlus, Users } from 'lucide-react'
import { toast } from 'sonner'
import { AdminPermission, hasAdminPermission } from '../access'
import type { AdminTextKey } from '../i18n'

type Text = (key: AdminTextKey) => string

interface CategoriesManagementPageProps {
  api: GooseAdminApi
  permissions: number[]
  text: Text
}

const emptyCategory: AdminCategory = { id: 0, category: '', desc: '', icon: '', color: '#3b82f6', slug: '', sort: 0 }

export function CategoriesManagementPage({ api, permissions, text }: CategoriesManagementPageProps) {
  const canManageCategories = hasAdminPermission(permissions, AdminPermission.TopicsManager)
  const canManageAccess = hasAdminPermission(permissions, AdminPermission.RoleManager)
  const [categories, setCategories] = useState<AdminCategory[]>([])
  const [overview, setOverview] = useState<AccessControlOverview>({ groups: [], categories: [] })
  const [globalModerators, setGlobalModerators] = useState<AdminCategoryModerator[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editor, setEditor] = useState<AdminCategory | null>(null)
  const [moderatorCategory, setModeratorCategory] = useState<AdminCategory | null>(null)
  const [accessCategory, setAccessCategory] = useState<AdminCategory | null>(null)
  const [globalOpen, setGlobalOpen] = useState(false)

  const beginRequest = useLatestRequest();
  const load = useCallback(async () => {
    const isCurrent = beginRequest();
    setLoading(true)
    setError('')
    try {
      const [categoryResult, accessResult, moderatorResult] = await Promise.all([
        canManageCategories ? api.categories.list() : Promise.resolve([]),
        canManageAccess ? api.categories.access() : Promise.resolve({ groups: [], categories: [] }),
        canManageCategories ? api.moderators.list() : Promise.resolve([]),
      ])
      const accessOnlyCategories = accessResult.categories.map((item) => ({
        id: item.id, category: item.name, color: item.color,
      }))
      if (!isCurrent()) return;
      setCategories(canManageCategories ? categoryResult : accessOnlyCategories)
      setOverview(accessResult)
      setGlobalModerators(moderatorResult)
    } catch (reason) {
      if (!isCurrent()) return;
      setError(errorMessage(reason, text('loadFailed')))
    } finally {
      if (isCurrent()) setLoading(false)
    }
  }, [beginRequest, api, canManageAccess, canManageCategories, text])

  useEffect(() => { void load() }, [load])
  useEffect(() => {
    if (!moderatorCategory) return
    const current = categories.find((item) => item.id === moderatorCategory.id)
    if (current && current !== moderatorCategory) setModeratorCategory(current)
  }, [categories, moderatorCategory])

  const filteredCategories = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase()
    if (!needle) return categories
    return categories.filter((item) => [item.category, item.slug, item.desc]
      .some((value) => value?.toLocaleLowerCase().includes(needle)))
  }, [categories, query])

  async function deleteCategory(category: AdminCategory) {
    if (!window.confirm(text('confirmDelete'))) return
    try {
      await api.categories.delete(category.id)
      toast.success(text('categoryDeleted'))
      await load()
    } catch (reason) {
      toast.error(errorMessage(reason, text('saveFailed')))
    }
  }

  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-semibold tracking-tight">{text('categories')}</h2>
            <Badge variant="outline">{categories.length}</Badge>
          </div>
          <p className="text-xs text-muted-foreground">{text('manageHint')}</p>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <div className="relative min-w-44 flex-1 sm:w-52 sm:flex-none">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={text('search')} className="h-7 pl-8 text-xs" />
          </div>
          {canManageCategories && (
            <Button variant="outline" size="sm" onClick={() => setGlobalOpen(true)}>
              <Users data-icon="inline-start" /> {text('globalModerators')} <Badge variant="secondary">{globalModerators.length}</Badge>
            </Button>
          )}
          {canManageCategories && <Button size="sm" onClick={() => setEditor({ ...emptyCategory })}><Plus data-icon="inline-start" /> {text('addCategory')}</Button>}
        </div>
      </header>

      {error && <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('loadFailed')}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert>}

      <section className="overflow-hidden border-y bg-background md:rounded-lg md:border">
          {loading ? (
            <div className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="size-4 animate-spin" />{text('loading')}</div>
          ) : filteredCategories.length === 0 ? (
            <Empty className="min-h-32 rounded-none border-0 py-4">
              <EmptyHeader>
                <EmptyMedia variant="icon"><Search /></EmptyMedia>
                <EmptyTitle>{text('empty')}</EmptyTitle>
                <EmptyDescription>{query ? text('search') : text('manageHint')}</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <>
              <div className="hidden md:block">
              <Table className="min-w-240 table-fixed">
                <TableHeader className="bg-muted/30"><TableRow>
                  <TableHead className="h-8 w-14">ID</TableHead>
                  <TableHead className="h-8 w-40">{text('name')}</TableHead>
                  {canManageAccess ? <TableHead className="h-8 w-28">{text('visibility')}</TableHead> : null}
                  {canManageCategories ? <TableHead className="h-8 w-28">Slug</TableHead> : null}
                  {canManageCategories ? <TableHead className="h-8">{text('description')}</TableHead> : null}
                  {canManageCategories ? <TableHead className="h-8 w-40">{text('moderators')}</TableHead> : null}
                  {canManageCategories ? <TableHead className="h-8 w-16">{text('sort')}</TableHead> : null}
                  <TableHead className="h-8 w-24 text-right">{text('actions')}</TableHead>
                </TableRow></TableHeader>
                <TableBody>{filteredCategories.map((category) => {
                  const access = overview.categories.find((item) => item.id === category.id)
                  return <TableRow key={category.id}>
                    <TableCell className="py-2 font-mono text-xs text-muted-foreground">{category.id}</TableCell>
                    <TableCell className="py-2"><div className="flex min-w-0 items-center gap-2">{category.icon ? <span style={{ color: category.color || undefined }}>{category.icon}</span> : <span className="size-2.5 shrink-0 rounded-[3px]" style={{ backgroundColor: category.color || '#64748b' }} />}<span className="truncate font-medium">{category.category}</span></div></TableCell>
                    {canManageAccess ? <TableCell className="py-2"><Badge variant={access?.isRestricted ? 'secondary' : 'outline'}>{access?.isRestricted ? <LockKeyhole data-icon="inline-start" /> : null}{access?.isRestricted ? text('restricted') : text('unrestricted')}</Badge></TableCell> : null}
                    {canManageCategories ? <TableCell className="truncate py-2 text-muted-foreground">{category.slug || '—'}</TableCell> : null}
                    {canManageCategories ? <TableCell className="truncate py-2 text-muted-foreground">{category.desc || '—'}</TableCell> : null}
                    {canManageCategories ? <TableCell className="py-2"><Button variant="outline" size="sm" onClick={() => setModeratorCategory(category)}><Shield data-icon="inline-start" />{category.moderators?.length ? `${text('moderators')} ${category.moderators.length}` : text('setModerator')}</Button></TableCell> : null}
                    {canManageCategories ? <TableCell className="py-2">{category.sort ?? 0}</TableCell> : null}
                    <TableCell className="py-2"><CategoryActions category={category} canManageCategories={canManageCategories} canManageAccess={canManageAccess} text={text} onEdit={() => setEditor({ ...category })} onDelete={() => void deleteCategory(category)} onAccess={() => setAccessCategory(category)} /></TableCell>
                  </TableRow>
                })}</TableBody>
              </Table>
              </div>
              <div className="divide-y md:hidden">
                {filteredCategories.map((category) => {
                  const access = overview.categories.find((item) => item.id === category.id)
                  return <article key={category.id} className="flex flex-col gap-2 px-3 py-2.5">
                    <div className="flex items-start justify-between gap-2">
                      <CategoryIdentity category={category} text={text} />
                      {canManageCategories ? <div className="flex shrink-0 gap-0.5"><Button variant="ghost" size="icon-sm" aria-label={text('edit')} onClick={() => setEditor({ ...category })}><Pencil /></Button><Button variant="ghost" size="icon-sm" aria-label={text('remove')} onClick={() => void deleteCategory(category)}><Trash2 /></Button></div> : null}
                    </div>
                    <div className="flex flex-wrap gap-1.5 pl-5">
                      <Button variant="outline" size="sm" disabled={!canManageCategories} onClick={() => setModeratorCategory(category)}><Users data-icon="inline-start" />{text('moderators')} · {category.moderators?.length || 0}</Button>
                      <Button variant="outline" size="sm" disabled={!canManageAccess} onClick={() => setAccessCategory(category)}><Shield data-icon="inline-start" />{access?.isRestricted ? text('restricted') : text('unrestricted')}</Button>
                    </div>
                  </article>
                })}
              </div>
            </>
          )}
      </section>

      <CategoryEditor category={editor} api={api} text={text} onClose={() => setEditor(null)} onSaved={async () => { setEditor(null); await load() }} />
      <ModeratorSheet open={globalOpen} title={text('globalModerators')} description={text('globalHint')} moderators={globalModerators} api={api} text={text} onOpenChange={setGlobalOpen} onChanged={load} />
      <ModeratorSheet open={Boolean(moderatorCategory)} title={moderatorCategory?.category || text('moderators')} description={text('moderators')} moderators={moderatorCategory?.moderators || []} categoryId={moderatorCategory?.id} api={api} text={text} onOpenChange={(open) => !open && setModeratorCategory(null)} onChanged={load} />
      <AccessSheet category={accessCategory} overview={overview} api={api} text={text} onClose={() => setAccessCategory(null)} onSaved={load} />
    </main>
  )
}

function CategoryActions({ category, canManageCategories, canManageAccess, text, onEdit, onDelete, onAccess }: {
  category: AdminCategory
  canManageCategories: boolean
  canManageAccess: boolean
  text: Text
  onEdit(): void
  onDelete(): void
  onAccess(): void
}) {
  return <div className="flex items-center justify-end gap-0.5">
    {canManageAccess ? <Button variant="ghost" size="icon-sm" title={text('access')} aria-label={`${text('access')} · ${category.category}`} onClick={onAccess}><LockKeyhole /></Button> : null}
    {canManageCategories && canManageAccess ? <DropdownMenu>
      <DropdownMenuTrigger asChild><Button variant="ghost" size="icon-sm" title={text('actions')} aria-label={`${text('actions')} · ${category.category}`}><MoreHorizontal /></Button></DropdownMenuTrigger>
      <DropdownMenuContent align="end"><DropdownMenuGroup>
        <DropdownMenuItem onSelect={onEdit}><Pencil />{text('edit')}</DropdownMenuItem>
        <DropdownMenuItem variant="destructive" onSelect={onDelete}><Trash2 />{text('remove')}</DropdownMenuItem>
      </DropdownMenuGroup></DropdownMenuContent>
    </DropdownMenu> : null}
    {canManageCategories && !canManageAccess ? <><Button variant="ghost" size="icon-sm" title={text('edit')} aria-label={`${text('edit')} · ${category.category}`} onClick={onEdit}><Pencil /></Button><Button variant="ghost" size="icon-sm" title={text('remove')} aria-label={`${text('remove')} · ${category.category}`} onClick={onDelete}><Trash2 /></Button></> : null}
  </div>
}

function CategoryIdentity({ category, text }: { category: AdminCategory; text: Text }) {
  return <div className="flex min-w-0 items-start gap-2">
    <span className="mt-1 size-2.5 shrink-0 rounded-full ring-1 ring-foreground/10" style={{ backgroundColor: category.color || 'var(--muted)' }} />
    <div className="min-w-0">
      <div className="truncate font-medium">{category.icon ? <span className="mr-1">{category.icon}</span> : null}{category.category}</div>
      {category.desc ? <div className="line-clamp-1 text-xs text-muted-foreground">{category.desc}</div> : null}
      <div className="mt-1 flex flex-wrap gap-1">
        <Badge variant="outline" className="font-mono font-normal">{category.slug || '—'}</Badge>
        <Badge variant="secondary">{text('sort')} {category.sort ?? 0}</Badge>
      </div>
    </div>
  </div>
}

function CategoryEditor({ category, api, text, onClose, onSaved }: { category: AdminCategory | null; api: GooseAdminApi; text: Text; onClose(): void; onSaved(): Promise<void> }) {
  const [form, setForm] = useState(emptyCategory)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  useEffect(() => { if (category) { setForm({ ...emptyCategory, ...category }); setError('') } }, [category])

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!form.category.trim()) { setError(text('categoryRequired')); return }
    setSaving(true); setError('')
    try {
      await api.categories.save({ ...form, category: form.category.trim(), slug: form.slug?.trim() })
      toast.success(text('categorySaved'))
      await onSaved()
    } catch (reason) { setError(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }

  return <Sheet open={Boolean(category)} onOpenChange={(open) => !open && onClose()}><SheetContent className="w-full overflow-y-auto sm:max-w-md">
    <SheetHeader><SheetTitle>{category?.id ? text('edit') : text('addCategory')}</SheetTitle><SheetDescription>{text('categoryFormHint')}</SheetDescription></SheetHeader>
    <form id="category-form" onSubmit={submit} className="px-4"><FieldGroup>
      <Field data-invalid={Boolean(error)}><FieldLabel htmlFor="category-name">{text('name')}</FieldLabel><Input id="category-name" autoFocus value={form.category} onChange={(event) => setForm({ ...form, category: event.target.value })} /><FieldError>{error}</FieldError></Field>
      <Field><FieldLabel htmlFor="category-slug">{text('slug')}</FieldLabel><Input id="category-slug" value={form.slug || ''} onChange={(event) => setForm({ ...form, slug: event.target.value })} /></Field>
      <Field><FieldLabel htmlFor="category-description">{text('description')}</FieldLabel><Input id="category-description" value={form.desc || ''} onChange={(event) => setForm({ ...form, desc: event.target.value })} /></Field>
      <div className="grid grid-cols-2 gap-4"><Field><FieldLabel htmlFor="category-icon">{text('icon')}</FieldLabel><Input id="category-icon" value={form.icon || ''} onChange={(event) => setForm({ ...form, icon: event.target.value })} /></Field><Field><FieldLabel htmlFor="category-color">{text('color')}</FieldLabel><ColorPicker id="category-color" label={text('color')} className="w-full" value={form.color || '#3b82f6'} onChange={(color) => setForm({ ...form, color })} /></Field></div>
      <Field><FieldLabel htmlFor="category-sort">{text('sort')}</FieldLabel><Input id="category-sort" type="number" value={form.sort ?? 0} onChange={(event) => setForm({ ...form, sort: Number(event.target.value) || 0 })} /><FieldDescription>{text('categoryFormHint')}</FieldDescription></Field>
    </FieldGroup></form>
    <SheetFooter className="flex-row justify-end"><Button variant="outline" onClick={onClose}>{text('cancel')}</Button><Button form="category-form" type="submit" disabled={saving}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button></SheetFooter>
  </SheetContent></Sheet>
}

function ModeratorSheet({ open, title, description, moderators, categoryId, api, text, onOpenChange, onChanged }: { open: boolean; title: string; description: string; moderators: AdminCategoryModerator[]; categoryId?: number; api: GooseAdminApi; text: Text; onOpenChange(open: boolean): void; onChanged(): Promise<void> }) {
  const [query, setQuery] = useState('')
  const [users, setUsers] = useState<AdminUser[]>([])
  const [searching, setSearching] = useState(false)
  useEffect(() => {
    if (!open || query.trim().length < 2) { setUsers([]); return }
    let active = true
    const timer = window.setTimeout(() => void searchUsers(), 220)
    async function searchUsers() {
      setSearching(true)
      const value = query.trim()
      const search = /^\d+$/.test(value) ? { userId: Number(value) } : { username: value }
      try { const result = await api.users.list({ page: 1, pageSize: 8, ...search }); if (active) setUsers(result.list) } catch { if (active) setUsers([]) } finally { if (active) setSearching(false) }
    }
    return () => { active = false; window.clearTimeout(timer) }
  }, [api, open, query])

  async function add(user: AdminUser) {
    try {
      if (categoryId) await api.categories.addModerator(categoryId, { userId: user.userId })
      else await api.moderators.add({ userId: user.userId })
      toast.success(text('moderatorAdded')); setQuery(''); setUsers([]); await onChanged()
    } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) }
  }
  async function remove(item: AdminCategoryModerator) {
    if (!window.confirm(text('confirmRemove'))) return
    try {
      if (categoryId) await api.categories.deleteModerator(item.id)
      else await api.moderators.delete(item.id)
      toast.success(text('moderatorRemoved')); await onChanged()
    } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) }
  }

  return <Sheet open={open} onOpenChange={onOpenChange}><SheetContent className="w-full overflow-y-auto sm:max-w-md"><SheetHeader><SheetTitle>{title}</SheetTitle><SheetDescription>{description}</SheetDescription></SheetHeader><div className="flex flex-col gap-5 px-4">
    <Field><FieldLabel htmlFor={`moderator-search-${categoryId || 'global'}`}>{text('addModerator')}</FieldLabel><div className="relative"><Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><Input id={`moderator-search-${categoryId || 'global'}`} value={query} onChange={(event) => setQuery(event.target.value)} placeholder={text('searchUser')} className="pl-8" />{searching && <LoaderCircle className="absolute right-2.5 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground" />}</div></Field>
    {query.trim().length >= 2 && !searching && <div className="rounded-lg border p-1">{users.length ? users.map((user) => <button type="button" key={user.userId} onClick={() => void add(user)} className="flex w-full items-center gap-3 rounded-md p-2 text-left hover:bg-muted"><UserAvatar user={user} /><span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium">{user.username}</span><span className="block truncate text-xs text-muted-foreground">{user.email}</span></span><UserPlus className="size-4" /></button>) : <p className="p-3 text-center text-sm text-muted-foreground">{text('noUsers')}</p>}</div>}
    <div className="flex flex-col gap-2">{moderators.map((item) => <div key={item.id} className="flex items-center gap-3 rounded-lg border p-2"><UserAvatar user={item} /><span className="min-w-0 flex-1 truncate text-sm font-medium">{item.username}</span><Badge variant={item.status === 1 ? 'secondary' : 'outline'}>{item.status === 1 ? text('active') : text('disabled')}</Badge><Button variant="ghost" size="icon-sm" aria-label={text('remove')} onClick={() => void remove(item)}><Trash2 /></Button></div>)}</div>
  </div></SheetContent></Sheet>
}

function AccessSheet({ category, overview, api, text, onClose, onSaved }: { category: AdminCategory | null; overview: AccessControlOverview; api: GooseAdminApi; text: Text; onClose(): void; onSaved(): Promise<void> }) {
  const [levels, setLevels] = useState<Record<number, number>>({})
  const [saving, setSaving] = useState(false)
  const accessMeta = overview.categories.find((item) => item.id === category?.id)
  useEffect(() => {
    if (!category) return
    setLevels(Object.fromEntries(overview.groups.map((group) => [group.id, group.grants.find((grant) => grant.categoryId === category.id)?.level || 0])))
  }, [category, overview])
  const groups = [...overview.groups].sort((left, right) => groupPriority(left.systemKey) - groupPriority(right.systemKey) || left.name.localeCompare(right.name))
  const everyone = groups.find((group) => group.systemKey === 'everyone')
  const restricted = everyone ? (levels[everyone.id] || 0) < 1 : true
  const currentEveryoneLevel = everyone?.grants.find((grant) => grant.categoryId === category?.id)?.level || 0
  const becomingRestricted = Boolean(everyone && currentEveryoneLevel >= 1 && restricted)
  const noAudience = groups.filter((group) => group.status === 1).every((group) => (levels[group.id] || 0) < 1)
  const conflict = Boolean(becomingRestricted && accessMeta && accessMeta.multiCategoryTopicCount > 0)
  const dirty = groups.some((group) => group.status === 1 && (levels[group.id] || 0) !== (group.grants.find((grant) => grant.categoryId === category?.id)?.level || 0))
  async function save() {
    if (!category || conflict) return
    setSaving(true)
    try {
      await api.categories.saveAccess(category.id, groups.filter((group) => group.status === 1).map((group) => ({ accessGroupId: group.id, level: levels[group.id] || 0 })))
      toast.success(text('accessSaved')); onClose(); await onSaved()
    } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  return <Sheet open={Boolean(category)} onOpenChange={(open) => !open && onClose()}><SheetContent className="w-full overflow-y-auto sm:max-w-lg"><SheetHeader><SheetTitle>{category?.category} · {text('access')}</SheetTitle><SheetDescription>{text('accessHint')}</SheetDescription></SheetHeader><div className="flex flex-col gap-3 px-4">
    {conflict && <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('restricted')}</AlertTitle><AlertDescription>{text('accessConflict')}</AlertDescription></Alert>}
    {becomingRestricted && !conflict && <Alert><AlertTriangle /><AlertTitle>{text('restricted')}</AlertTitle><AlertDescription>{text('legacyImages')}</AlertDescription></Alert>}
    {noAudience && <Alert variant="destructive"><AlertTriangle /><AlertDescription>{text('noAudience')}</AlertDescription></Alert>}
    {groups.map((group) => <div key={group.id} className={`flex items-center justify-between gap-4 rounded-lg border p-3 ${group.status !== 1 ? 'opacity-55' : ''}`}><div className="min-w-0"><div className="truncate text-sm font-medium">{group.name || (group.systemKey === 'everyone' ? text('everyone') : group.systemKey)}</div><div className="mt-0.5 text-xs text-muted-foreground">{group.systemKey || group.joinMode}</div></div><Select disabled={group.status !== 1} value={String(levels[group.id] || 0)} onValueChange={(value) => setLevels({ ...levels, [group.id]: Number(value) })}><SelectTrigger className="w-36"><SelectValue /></SelectTrigger><SelectContent>{[[0, 'none'], [1, 'read'], [2, 'reply'], [3, 'create'], [4, 'manage']].map(([level, key]) => <SelectItem key={level} value={String(level)}>{text(key as AdminTextKey)}</SelectItem>)}</SelectContent></Select></div>)}
  </div><SheetFooter className="flex-row justify-end"><Button variant="outline" onClick={onClose}>{text('cancel')}</Button><Button disabled={saving || conflict || !everyone || !dirty} onClick={() => void save()}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button></SheetFooter></SheetContent></Sheet>
}

function UserAvatar({ user }: { user: { username: string; avatarUrl?: string | null } }) {
  return <Avatar className="size-8"><AvatarImage src={user.avatarUrl || undefined} alt={user.username} /><AvatarFallback>{user.username.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar>
}

function errorMessage(reason: unknown, fallback: string) {
  return reason instanceof Error && reason.message ? reason.message : fallback
}

function groupPriority(systemKey?: string) {
  if (systemKey === 'everyone') return 0
  if (systemKey === 'registered') return 1
  return 2
}
