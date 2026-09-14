import { useCallback, useEffect, useMemo, useState } from 'react'
import type { AdminPermissionOption, AdminRole, GooseAdminApi } from '@gooseforum/client'
import { Alert, AlertDescription, AlertTitle } from '@gooseforum/react/components/ui/alert'
import { Badge } from '@gooseforum/react/components/ui/badge'
import { Button } from '@gooseforum/react/components/ui/button'
import { Checkbox } from '@gooseforum/react/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/react/components/ui/dialog'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/react/components/ui/empty'
import { Field, FieldError, FieldGroup, FieldLabel, FieldLegend, FieldSet } from '@gooseforum/react/components/ui/field'
import { Input } from '@gooseforum/react/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/react/components/ui/select'
import { Spinner } from '@gooseforum/react/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/react/components/ui/table'
import { AlertTriangle, Pencil, Plus, RefreshCw, Search, ShieldCheck, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import type { RoleTextKey } from '../roles-i18n'

type Text = (key: RoleTextKey) => string
type RoleDraft = { id: number; roleName: string; permissions: number[] }
const emptyDraft: RoleDraft = { id: 0, roleName: '', permissions: [] }

export function RolesManagementPage({ api, text }: { api: GooseAdminApi; text: Text }) {
  const [roles, setRoles] = useState<AdminRole[]>([])
  const [permissionOptions, setPermissionOptions] = useState<AdminPermissionOption[]>([])
  const [search, setSearch] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [effective, setEffective] = useState('all')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [draft, setDraft] = useState<RoleDraft | null>(null)
  const [deletingRole, setDeletingRole] = useState<AdminRole | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [roleResult, permissions] = await Promise.all([api.roles.list(), api.roles.permissions()])
      setRoles(roleResult.list || [])
      setPermissionOptions(permissions)
    } catch (reason) {
      setError(errorMessage(reason, text('loadFailed')))
    } finally {
      setLoading(false)
    }
  }, [api, text])

  useEffect(() => { void load() }, [load])

  const filteredRoles = useMemo(() => {
    const needle = appliedSearch.toLocaleLowerCase()
    return roles.filter((role) => (!needle || role.roleName.toLocaleLowerCase().includes(needle)) && (effective === 'all' || String(role.effective) === effective))
  }, [appliedSearch, effective, roles])
  const pageCount = Math.max(1, Math.ceil(filteredRoles.length / pageSize))
  const currentPage = Math.min(page, pageCount)
  const visibleRoles = filteredRoles.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  function applySearch(event: React.FormEvent) {
    event.preventDefault()
    setAppliedSearch(search.trim())
    setPage(1)
  }

  return <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
    <header className="flex items-center justify-between gap-3"><div className="min-w-0"><h2 className="text-lg font-semibold tracking-tight">{text('title')}</h2><p className="truncate text-xs text-muted-foreground">{text('description')}</p></div><Button size="sm" onClick={() => setDraft({ ...emptyDraft })}><Plus data-icon="inline-start" />{text('create')}</Button></header>
    {error ? <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('loadFailed')}</AlertTitle><AlertDescription className="flex items-center justify-between gap-3"><span>{error}</span><Button variant="outline" size="sm" onClick={() => void load()}>{text('retry')}</Button></AlertDescription></Alert> : null}
    <section className="overflow-hidden rounded-lg border bg-background">
      <div className="flex flex-col gap-2 border-b bg-muted/20 p-2 sm:flex-row sm:items-center sm:justify-between">
        <form className="flex min-w-0 flex-1 gap-1.5 sm:max-w-sm" onSubmit={applySearch}><div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} className="h-8 pl-8" placeholder={text('search')} /></div><Button size="sm" type="submit">{text('searchAction')}</Button></form>
        <div className="flex items-center gap-1.5"><Select value={effective} onValueChange={(value) => { setEffective(value); setPage(1) }}><SelectTrigger size="sm" className="w-28"><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="all">{text('all')}</SelectItem><SelectItem value="1">{text('enabled')}</SelectItem><SelectItem value="0">{text('disabled')}</SelectItem></SelectGroup></SelectContent></Select><Button variant="outline" size="sm" disabled={loading} onClick={() => void load()}><RefreshCw data-icon="inline-start" className={loading ? 'animate-spin' : undefined} />{text('refresh')}</Button></div>
      </div>
      {loading && !roles.length ? <Empty className="min-h-32 rounded-none border-0"><EmptyHeader><EmptyMedia variant="icon"><Spinner /></EmptyMedia><EmptyTitle>{text('loading')}</EmptyTitle></EmptyHeader></Empty> : !visibleRoles.length ? <Empty className="min-h-32 rounded-none border-0"><EmptyHeader><EmptyMedia variant="icon"><Search /></EmptyMedia><EmptyTitle>{text('empty')}</EmptyTitle></EmptyHeader></Empty> : <Table className="min-w-200 table-fixed"><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8 w-16">ID</TableHead><TableHead className="h-8 w-40">{text('name')}</TableHead><TableHead className="h-8 w-24">{text('status')}</TableHead><TableHead className="h-8">{text('permissions')}</TableHead><TableHead className="h-8 w-44">{text('createdAt')}</TableHead><TableHead className="h-8 w-20 text-right">{text('actions')}</TableHead></TableRow></TableHeader><TableBody>{visibleRoles.map((role) => <TableRow key={role.roleId}><TableCell className="py-2 font-mono text-xs text-muted-foreground">{role.roleId}</TableCell><TableCell className="truncate py-2 font-medium">{role.roleName}</TableCell><TableCell className="py-2"><Badge variant={role.effective === 1 ? 'default' : 'secondary'}><ShieldCheck data-icon="inline-start" />{role.effective === 1 ? text('enabled') : text('disabled')}</Badge></TableCell><TableCell className="py-2"><div className="flex flex-wrap gap-1">{role.permissions.map((permission) => <Badge key={permission.id} variant="outline">{permission.name}</Badge>)}</div></TableCell><TableCell className="truncate py-2 text-xs text-muted-foreground">{role.createTime || '—'}</TableCell><TableCell className="py-2"><div className="flex justify-end gap-0.5"><Button variant="ghost" size="icon-sm" title={text('edit')} onClick={() => setDraft({ id: role.roleId, roleName: role.roleName, permissions: role.permissions.map((permission) => permission.id) })}><Pencil /></Button><Button variant="ghost" size="icon-sm" title={text('delete')} onClick={() => setDeletingRole(role)}><Trash2 /></Button></div></TableCell></TableRow>)}</TableBody></Table>}
      <footer className="flex flex-wrap items-center justify-between gap-2 border-t px-2 py-1.5 text-xs text-muted-foreground"><span>{filteredRoles.length} · {text('page')} {currentPage}/{pageCount}</span><div className="flex items-center gap-1"><Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1) }}><SelectTrigger size="sm" className="w-20"><SelectValue /></SelectTrigger><SelectContent><SelectGroup>{[10, 20, 50].map((size) => <SelectItem key={size} value={String(size)}>{size}</SelectItem>)}</SelectGroup></SelectContent></Select><Button variant="outline" size="sm" disabled={currentPage <= 1} onClick={() => setPage(currentPage - 1)}>{text('previous')}</Button><Button variant="outline" size="sm" disabled={currentPage >= pageCount} onClick={() => setPage(currentPage + 1)}>{text('next')}</Button></div></footer>
    </section>
    <RoleEditor value={draft} permissionOptions={permissionOptions} api={api} text={text} onClose={() => setDraft(null)} onSaved={async () => { setDraft(null); await load() }} />
    <DeleteRoleDialog role={deletingRole} api={api} text={text} onClose={() => setDeletingRole(null)} onDeleted={async () => { setDeletingRole(null); await load() }} />
  </main>
}

function RoleEditor({ value, permissionOptions, api, text, onClose, onSaved }: { value: RoleDraft | null; permissionOptions: AdminPermissionOption[]; api: GooseAdminApi; text: Text; onClose(): void; onSaved(): Promise<void> }) {
  const [form, setForm] = useState(emptyDraft)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  useEffect(() => { if (value) { setForm(value); setError('') } }, [value])
  function togglePermission(id: number, checked: boolean) { setForm((current) => ({ ...current, permissions: checked ? [...new Set([...current.permissions, id])] : current.permissions.filter((value) => value !== id) })) }
  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!form.roleName.trim()) { setError(text('nameRequired')); return }
    if (!form.permissions.length) { setError(text('permissionRequired')); return }
    setSaving(true); setError('')
    try { await api.roles.save({ ...form, roleName: form.roleName.trim() }); toast.success(text('saved')); await onSaved() } catch (reason) { setError(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  return <Dialog open={Boolean(value)} onOpenChange={(open) => !open && !saving && onClose()}><DialogContent className="sm:max-w-lg"><DialogHeader><DialogTitle>{form.id ? text('editTitle') : text('createTitle')}</DialogTitle><DialogDescription>{text('formHint')}</DialogDescription></DialogHeader><form id="role-form" onSubmit={submit}><FieldGroup><Field data-invalid={Boolean(error)}><FieldLabel htmlFor="role-name">{text('name')}</FieldLabel><Input id="role-name" autoFocus aria-invalid={Boolean(error)} value={form.roleName} onChange={(event) => setForm({ ...form, roleName: event.target.value })} placeholder={text('namePlaceholder')} /><FieldError>{error}</FieldError></Field><FieldSet><FieldLegend variant="label">{text('permissions')}</FieldLegend><div className="grid max-h-56 grid-cols-1 gap-2 overflow-y-auto rounded-lg border p-3 sm:grid-cols-2">{permissionOptions.map((permission) => <Field key={permission.value} orientation="horizontal"><Checkbox id={`permission-${permission.value}`} checked={form.permissions.includes(permission.value)} onCheckedChange={(checked) => togglePermission(permission.value, Boolean(checked))} /><FieldLabel htmlFor={`permission-${permission.value}`}>{permission.label || permission.name}</FieldLabel></Field>)}</div></FieldSet></FieldGroup></form><DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button form="role-form" type="submit" disabled={saving}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button></DialogFooter></DialogContent></Dialog>
}

function DeleteRoleDialog({ role, api, text, onClose, onDeleted }: { role: AdminRole | null; api: GooseAdminApi; text: Text; onClose(): void; onDeleted(): Promise<void> }) {
  const [saving, setSaving] = useState(false)
  async function remove() { if (!role) return; setSaving(true); try { await api.roles.delete(role.roleId); toast.success(text('deleted')); await onDeleted() } catch (reason) { toast.error(errorMessage(reason, text('deleteFailed'))) } finally { setSaving(false) } }
  return <Dialog open={Boolean(role)} onOpenChange={(open) => !open && !saving && onClose()}><DialogContent><DialogHeader><DialogTitle>{text('deleteTitle')}</DialogTitle><DialogDescription>{text('deleteConfirm')} {role?.roleName}</DialogDescription></DialogHeader><DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button variant="destructive" disabled={saving} onClick={() => void remove()}>{saving ? <Spinner data-icon="inline-start" /> : <Trash2 data-icon="inline-start" />}{text('delete')}</Button></DialogFooter></DialogContent></Dialog>
}

function errorMessage(reason: unknown, fallback: string) { return reason instanceof Error && reason.message ? reason.message : fallback }
