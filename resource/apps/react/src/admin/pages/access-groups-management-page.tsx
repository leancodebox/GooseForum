import { useCallback, useEffect, useMemo, useState } from 'react'
import type { AccessControlOverview, AccessGroup, AccessGroupMember, GooseAdminApi } from '@gooseforum/client'
import { Alert, AlertDescription, AlertTitle } from '@gooseforum/react/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@gooseforum/react/components/ui/avatar'
import { Badge } from '@gooseforum/react/components/ui/badge'
import { Button } from '@gooseforum/react/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/react/components/ui/dialog'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/react/components/ui/empty'
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@gooseforum/react/components/ui/field'
import { Input } from '@gooseforum/react/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/react/components/ui/select'
import { Spinner } from '@gooseforum/react/components/ui/spinner'
import { Switch } from '@gooseforum/react/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/react/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@gooseforum/react/components/ui/tabs'
import { AlertTriangle, Check, Clock3, LockKeyhole, Pencil, Plus, RefreshCw, Search, ShieldCheck, Trash2, UserPlus, UsersRound } from 'lucide-react'
import { toast } from 'sonner'
import type { AccessGroupTextKey } from '../access-groups-i18n'
import { shouldUseClientNavigation } from '../navigation'

type Text = (key: AccessGroupTextKey) => string
type GroupForm = { id: number; name: string; joinMode: 'invite_only' | 'application'; status: number }

const emptyOverview: AccessControlOverview = { groups: [], categories: [] }
const emptyGroupForm: GroupForm = { id: 0, name: '', joinMode: 'invite_only', status: 1 }

export function AccessGroupsManagementPage({ api, text, onNavigate }: { api: GooseAdminApi; text: Text; onNavigate(path: string): void }) {
  const [overview, setOverview] = useState<AccessControlOverview>(emptyOverview)
  const [selectedGroupId, setSelectedGroupId] = useState(0)
  const [panel, setPanel] = useState<'permissions' | 'members'>('permissions')
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [groupForm, setGroupForm] = useState<GroupForm | null>(null)
  const [deletingGroup, setDeletingGroup] = useState<AccessGroup | null>(null)
  const [deletingMember, setDeletingMember] = useState<AccessGroupMember | null>(null)

  const load = useCallback(async (preferredGroupId = 0) => {
    setLoading(true)
    setError('')
    try {
      const result = await api.accessGroups.overview()
      setOverview(result)
      const nextGroupId = result.groups.some((group) => group.id === preferredGroupId) ? preferredGroupId : result.groups[0]?.id || 0
      setSelectedGroupId(nextGroupId)
      if (result.groups.find((group) => group.id === nextGroupId)?.systemKey) setPanel('permissions')
    } catch (reason) {
      setError(errorMessage(reason, text('loadFailed')))
    } finally {
      setLoading(false)
    }
  }, [api, text])

  useEffect(() => { void load() }, [load])

  const selectedGroup = overview.groups.find((group) => group.id === selectedGroupId)
  const filteredGroups = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase()
    if (!needle) return overview.groups
    return overview.groups.filter((group) => `${group.name} ${group.systemKey || group.joinMode}`.toLocaleLowerCase().includes(needle))
  }, [overview.groups, query])
  const selectedMembers = useMemo(() => [...(selectedGroup?.members || [])].sort((left, right) => {
    if (left.status !== right.status) return right.status - left.status
    if (left.memberRole !== right.memberRole) return left.memberRole === 'manager' ? -1 : 1
    return left.userId - right.userId
  }), [selectedGroup])

  function selectGroup(group: AccessGroup) {
    setSelectedGroupId(group.id)
    setPanel('permissions')
  }

  return <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
    <header className="flex items-center justify-between gap-3">
      <div className="min-w-0"><h2 className="text-lg font-semibold tracking-tight">{text('title')}</h2><p className="truncate text-xs text-muted-foreground">{text('description')}</p></div>
      <Button size="sm" onClick={() => setGroupForm({ ...emptyGroupForm })}><Plus data-icon="inline-start" />{text('create')}</Button>
    </header>

    {error ? <Alert variant="destructive"><AlertTriangle /><AlertTitle>{text('loadFailed')}</AlertTitle><AlertDescription className="flex items-center justify-between gap-3"><span>{error}</span><Button variant="outline" size="sm" onClick={() => void load(selectedGroupId)}>{text('retry')}</Button></AlertDescription></Alert> : null}

    <div className="grid items-start gap-3 xl:grid-cols-[15rem_minmax(0,1fr)]">
      <section className="overflow-hidden rounded-lg border bg-background">
        <div className="flex flex-col gap-2 border-b p-3">
          <div className="flex items-center justify-between gap-2"><span className="flex items-center gap-2 text-sm font-semibold"><UsersRound className="size-4" />{text('groups')}</span><span className="flex items-center gap-1"><Badge variant="outline">{overview.groups.length}</Badge><Button variant="ghost" size="icon-sm" disabled={loading} title={text('refresh')} onClick={() => void load(selectedGroupId)}><RefreshCw className={loading ? 'animate-spin' : undefined} /></Button></span></div>
          <p className="text-xs text-muted-foreground">{text('systemGroupHint')}</p>
          <label className="relative"><Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input aria-label={text('groups')} value={query} onChange={(event) => setQuery(event.target.value)} className="h-8 pl-8 text-sm" placeholder={text('searchGroups')} /></label>
        </div>
        {loading && !overview.groups.length ? <GroupListEmpty icon={<Spinner />} title={text('loading')} /> : !filteredGroups.length ? <GroupListEmpty icon={<Search />} title={overview.groups.length ? text('noMatchingGroups') : text('noGroups')} /> : <nav className="divide-y" aria-label={text('groups')}>{filteredGroups.map((group) => <Button key={group.id} type="button" variant="ghost" data-disabled={group.status !== 1} aria-current={selectedGroupId === group.id ? 'true' : undefined} onClick={() => selectGroup(group)} className="h-auto w-full justify-start rounded-none px-3 py-2 text-left font-normal shadow-none hover:bg-muted/60 aria-[current=true]:bg-primary/10 aria-[current=true]:text-primary data-[disabled=true]:opacity-60">
          <span className="grid size-7 shrink-0 place-items-center rounded-md border bg-background">{group.systemKey ? <ShieldCheck className="size-3.5" /> : <UsersRound className="size-3.5" />}</span>
          <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium">{group.name}</span><span className="block truncate text-[11px] text-muted-foreground">{groupTypeLabel(group, text)}</span></span>
          <Badge variant="outline">{activeMemberCount(group)}</Badge>
        </Button>)}</nav>}
      </section>

      {selectedGroup ? <GroupDetails group={selectedGroup} members={selectedMembers} overview={overview} panel={panel} api={api} text={text} loading={loading} onPanelChange={setPanel} onEdit={() => setGroupForm({ id: selectedGroup.id, name: selectedGroup.name, joinMode: selectedGroup.joinMode === 'application' ? 'application' : 'invite_only', status: selectedGroup.status })} onDelete={() => setDeletingGroup(selectedGroup)} onDeleteMember={setDeletingMember} onReload={load} onNavigate={onNavigate} /> : <section className="rounded-lg border bg-background"><Empty className="min-h-40 border-0"><EmptyHeader><EmptyMedia variant="icon"><UsersRound /></EmptyMedia><EmptyTitle>{loading ? text('loading') : text('selectGroup')}</EmptyTitle></EmptyHeader></Empty></section>}
    </div>

    <GroupEditor value={groupForm} api={api} text={text} onClose={() => setGroupForm(null)} onSaved={async (id) => { setGroupForm(null); await load(id) }} />
    <ConfirmDialog open={Boolean(deletingGroup)} title={text('deleteGroupTitle')} description={text('deleteGroupConfirm')} confirmText={text('delete')} apiCall={async () => { if (deletingGroup) await api.accessGroups.delete(deletingGroup.id) }} onClose={() => setDeletingGroup(null)} onDone={async () => { toast.success(text('groupDeleted')); setDeletingGroup(null); await load(0) }} text={text} />
    <ConfirmDialog open={Boolean(deletingMember)} title={text('removeMemberTitle')} description={text('removeMemberConfirm')} confirmText={text('removeMember')} apiCall={async () => { if (deletingMember && selectedGroup) await api.accessGroups.deleteMember(selectedGroup.id, deletingMember.id) }} onClose={() => setDeletingMember(null)} onDone={async () => { toast.success(text('memberDeleted')); setDeletingMember(null); await load(selectedGroupId) }} text={text} />
  </main>
}

function GroupDetails({ group, members, overview, panel, api, text, loading, onPanelChange, onEdit, onDelete, onDeleteMember, onReload, onNavigate }: { group: AccessGroup; members: AccessGroupMember[]; overview: AccessControlOverview; panel: 'permissions' | 'members'; api: GooseAdminApi; text: Text; loading: boolean; onPanelChange(value: 'permissions' | 'members'): void; onEdit(): void; onDelete(): void; onDeleteMember(member: AccessGroupMember): void; onReload(id?: number): Promise<void>; onNavigate(path: string): void }) {
  const pendingCount = group.members.filter((member) => member.status === 2).length
  return <section className="min-w-0 rounded-lg border bg-background p-3">
    <div className="flex flex-col gap-2 border-b pb-3 sm:flex-row sm:items-start sm:justify-between">
      <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><h2 className="truncate text-lg font-semibold">{group.name}</h2>{group.systemKey ? <Badge variant="secondary"><ShieldCheck data-icon="inline-start" />{text('systemGroup')}</Badge> : null}<Badge variant={group.status === 1 ? 'outline' : 'destructive'}>{group.status === 1 ? text('enabled') : text('disabled')}</Badge></div>
        <p className="mt-1 text-sm text-muted-foreground">{group.systemKey ? text('systemImmutable') : group.joinMode === 'application' ? text('joinApplicationDescription') : text('joinInviteDescription')}</p>
        <div className="mt-1.5 flex flex-wrap gap-3 text-xs text-muted-foreground"><span>{activeMemberCount(group)} {text('memberCount')}</span>{pendingCount ? <span>{pendingCount} {text('pendingCount')}</span> : null}<span>{group.grants.filter((grant) => grant.level > 0).length} {text('grantCount')}</span><span className="font-mono">ID {group.id}</span></div>
      </div>
      {!group.systemKey ? <div className="flex shrink-0 gap-1"><Button variant="outline" size="sm" onClick={onEdit}><Pencil data-icon="inline-start" />{text('edit')}</Button><Button variant="destructive" size="sm" onClick={onDelete}><Trash2 data-icon="inline-start" />{text('delete')}</Button></div> : null}
    </div>

    <Tabs value={panel} onValueChange={(value) => onPanelChange(value as 'permissions' | 'members')} className="mt-3 min-w-0 gap-3">
      <TabsList><TabsTrigger value="permissions"><LockKeyhole />{text('categoryPermissions')}</TabsTrigger>{!group.systemKey ? <TabsTrigger value="members"><UsersRound />{text('members')}{pendingCount ? <Badge variant="secondary">{pendingCount}</Badge> : null}</TabsTrigger> : null}</TabsList>
      <TabsContent value="permissions"><PermissionsPanel group={group} overview={overview} text={text} onNavigate={onNavigate} /></TabsContent>
      {!group.systemKey ? <TabsContent value="members"><MembersPanel group={group} members={members} api={api} text={text} loading={loading} onDelete={onDeleteMember} onReload={onReload} /></TabsContent> : null}
    </Tabs>
  </section>
}

function PermissionsPanel({ group, overview, text, onNavigate }: { group: AccessGroup; overview: AccessControlOverview; text: Text; onNavigate(path: string): void }) {
  return <div className="flex flex-col gap-2"><div className="flex items-center justify-between gap-3"><p className="text-xs text-muted-foreground">{text('permissionSummaryHint')}</p><Button variant="outline" size="sm" asChild><a href="/admin/categories" onClick={(event) => { if (!shouldUseClientNavigation(event)) return; event.preventDefault(); onNavigate('/admin/categories') }}><Pencil data-icon="inline-start" />{text('manageCategoryPermissions')}</a></Button></div>
    <div className="overflow-hidden rounded-lg border"><Table><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8">{text('category')}</TableHead><TableHead className="h-8">{text('visibility')}</TableHead><TableHead className="h-8 w-40">{text('capability')}</TableHead></TableRow></TableHeader><TableBody>{overview.categories.length ? overview.categories.map((category) => <TableRow key={category.id}><TableCell className="py-2"><span className="flex items-center gap-2 font-medium"><span className="size-2.5 rounded-[3px]" style={{ backgroundColor: category.color || '#64748b' }} />{category.name}</span></TableCell><TableCell className="py-2"><Badge variant={category.isRestricted ? 'secondary' : 'outline'}>{category.isRestricted ? <LockKeyhole data-icon="inline-start" /> : <Check data-icon="inline-start" />}{category.isRestricted ? text('restricted') : text('public')}</Badge></TableCell><TableCell className="py-2"><Badge variant="outline">{levelLabel(grantLevel(group, category.id), text)}</Badge></TableCell></TableRow>) : <TableRow><TableCell colSpan={3} className="h-24 text-center text-muted-foreground">{text('category')}</TableCell></TableRow>}</TableBody></Table></div>
  </div>
}

function MembersPanel({ group, members, api, text, loading, onDelete, onReload }: { group: AccessGroup; members: AccessGroupMember[]; api: GooseAdminApi; text: Text; loading: boolean; onDelete(member: AccessGroupMember): void; onReload(id?: number): Promise<void> }) {
  const [username, setUsername] = useState('')
  const [role, setRole] = useState<'member' | 'manager'>('member')
  const [saving, setSaving] = useState(false)
  const [reviewing, setReviewing] = useState(0)
  async function addMember(event: React.FormEvent) {
    event.preventDefault()
    const value = username.trim()
    if (!value || saving) return
    setSaving(true)
    try { await api.accessGroups.saveMember({ groupId: group.id, memberRole: role, ...(/^\d+$/.test(value) ? { userId: Number(value) } : { username: value }) }); setUsername(''); toast.success(text('memberSaved')); await onReload(group.id) } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  async function review(memberId: number, approve: boolean) {
    if (reviewing) return
    setReviewing(memberId)
    try { await api.accessGroups.reviewApplication(group.id, memberId, approve); toast.success(text(approve ? 'applicationApproved' : 'applicationRejected')); await onReload(group.id) } catch (reason) { toast.error(errorMessage(reason, text('saveFailed'))) } finally { setReviewing(0) }
  }
  return <div className="flex flex-col gap-2"><p className="text-xs text-muted-foreground">{text('memberHint')}</p><form onSubmit={addMember} className="grid gap-2 rounded-lg border bg-muted/20 p-2 lg:grid-cols-[minmax(0,1fr)_9rem_auto] lg:items-end"><Field><FieldLabel htmlFor="access-member">{text('username')}</FieldLabel><Input id="access-member" value={username} onChange={(event) => setUsername(event.target.value)} disabled={saving || group.status !== 1} placeholder={text('usernamePlaceholder')} /></Field><Field><FieldLabel>{text('role')}</FieldLabel><Select value={role} disabled={saving || group.status !== 1} onValueChange={(value) => setRole(value as 'member' | 'manager')}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="member">{text('roleMember')}</SelectItem><SelectItem value="manager">{text('roleManager')}</SelectItem></SelectGroup></SelectContent></Select></Field><Button type="submit" disabled={saving || loading || group.status !== 1 || !username.trim()}>{saving ? <Spinner data-icon="inline-start" /> : <UserPlus data-icon="inline-start" />}{text('addMember')}</Button></form>
    <div className="overflow-hidden rounded-lg border"><Table><TableHeader className="bg-muted/30"><TableRow><TableHead className="h-8">{text('member')}</TableHead><TableHead className="h-8">{text('role')}</TableHead><TableHead className="h-8">{text('status')}</TableHead><TableHead className="h-8 w-40 text-right">{text('actions')}</TableHead></TableRow></TableHeader><TableBody>{members.length ? members.map((member) => <TableRow key={member.id}><TableCell className="py-2"><div className="flex items-center gap-2"><Avatar className="size-8"><AvatarImage src={member.avatarUrl || undefined} alt={member.username} /><AvatarFallback>{memberInitial(member)}</AvatarFallback></Avatar><span className="min-w-0"><span className="block truncate font-medium">{member.username || `#${member.userId}`}</span><span className="block font-mono text-xs text-muted-foreground">ID {member.userId}</span></span></div></TableCell><TableCell className="py-2"><Badge variant="outline">{member.memberRole === 'manager' ? text('roleManager') : text('roleMember')}</Badge></TableCell><TableCell className="py-2"><Badge variant={member.status === 2 ? 'secondary' : 'outline'}>{member.status === 2 ? <Clock3 data-icon="inline-start" /> : <Check data-icon="inline-start" />}{member.status === 2 ? text('pending') : text('active')}</Badge></TableCell><TableCell className="py-2"><div className="flex justify-end gap-1">{member.status === 2 ? <><Button variant="outline" size="sm" disabled={Boolean(reviewing)} onClick={() => void review(member.id, false)}>{text('reject')}</Button><Button size="sm" disabled={Boolean(reviewing)} onClick={() => void review(member.id, true)}>{text('approve')}</Button></> : <Button variant="ghost" size="icon-sm" title={text('removeMember')} onClick={() => onDelete(member)}><Trash2 /></Button>}</div></TableCell></TableRow>) : <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">{text('noMembers')}</TableCell></TableRow>}</TableBody></Table></div>
  </div>
}

function GroupEditor({ value, api, text, onClose, onSaved }: { value: GroupForm | null; api: GooseAdminApi; text: Text; onClose(): void; onSaved(id: number): Promise<void> }) {
  const [form, setForm] = useState(emptyGroupForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  useEffect(() => { if (value) { setForm(value); setError('') } }, [value])
  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!form.name.trim()) { setError(text('groupNameRequired')); return }
    setSaving(true)
    try { const id = await api.accessGroups.save({ ...form, name: form.name.trim() }); toast.success(text('groupSaved')); await onSaved(id) } catch (reason) { setError(errorMessage(reason, text('saveFailed'))) } finally { setSaving(false) }
  }
  return <Dialog open={Boolean(value)} onOpenChange={(open) => !open && !saving && onClose()}><DialogContent><DialogHeader><DialogTitle>{form.id ? text('editGroup') : text('createGroup')}</DialogTitle><DialogDescription>{text('groupFormHint')}</DialogDescription></DialogHeader><form id="access-group-form" onSubmit={submit}><FieldGroup><Field data-invalid={Boolean(error)}><FieldLabel htmlFor="access-group-name">{text('groupName')}</FieldLabel><Input id="access-group-name" autoFocus aria-invalid={Boolean(error)} value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} placeholder={text('groupNamePlaceholder')} /><FieldError>{error}</FieldError></Field><Field><FieldLabel>{text('joinMode')}</FieldLabel><Select value={form.joinMode} onValueChange={(value) => setForm({ ...form, joinMode: value as GroupForm['joinMode'] })}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="invite_only">{text('inviteOnly')}</SelectItem><SelectItem value="application">{text('application')}</SelectItem></SelectGroup></SelectContent></Select></Field><Field orientation="horizontal"><div className="flex-1"><FieldLabel>{text('enabled')}</FieldLabel><FieldDescription>{text('enabledHint')}</FieldDescription></div><Switch checked={form.status === 1} onCheckedChange={(checked) => setForm({ ...form, status: checked ? 1 : 0 })} /></Field></FieldGroup></form><DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button form="access-group-form" type="submit" disabled={saving}>{saving ? <Spinner data-icon="inline-start" /> : null}{saving ? text('saving') : text('save')}</Button></DialogFooter></DialogContent></Dialog>
}

function ConfirmDialog({ open, title, description, confirmText, apiCall, onClose, onDone, text }: { open: boolean; title: string; description: string; confirmText: string; apiCall(): Promise<void>; onClose(): void; onDone(): Promise<void>; text: Text }) {
  const [saving, setSaving] = useState(false)
  async function confirm() { setSaving(true); try { await apiCall(); await onDone() } catch (reason) { toast.error(errorMessage(reason, text('deleteFailed'))) } finally { setSaving(false) } }
  return <Dialog open={open} onOpenChange={(next) => !next && !saving && onClose()}><DialogContent><DialogHeader><DialogTitle>{title}</DialogTitle><DialogDescription>{description}</DialogDescription></DialogHeader><DialogFooter><Button variant="outline" disabled={saving} onClick={onClose}>{text('cancel')}</Button><Button variant="destructive" disabled={saving} onClick={() => void confirm()}>{saving ? <Spinner data-icon="inline-start" /> : <Trash2 data-icon="inline-start" />}{confirmText}</Button></DialogFooter></DialogContent></Dialog>
}

function GroupListEmpty({ icon, title }: { icon: React.ReactNode; title: string }) { return <Empty className="min-h-28 rounded-none border-0 p-4"><EmptyHeader><EmptyMedia variant="icon">{icon}</EmptyMedia><EmptyTitle>{title}</EmptyTitle></EmptyHeader></Empty> }
function activeMemberCount(group: AccessGroup) { return group.members.filter((member) => member.status === 1).length }
function grantLevel(group: AccessGroup, categoryId: number) { return group.grants.find((grant) => grant.categoryId === categoryId)?.level || 0 }
function levelLabel(level: number, text: Text) { return text((['none', 'read', 'reply', 'createTopics', 'manage'][level] || 'none') as AccessGroupTextKey) }
function memberInitial(member: AccessGroupMember) { return (member.username || String(member.userId)).slice(0, 1).toUpperCase() }
function groupTypeLabel(group: AccessGroup, text: Text) { if (group.systemKey === 'everyone') return text('systemEveryone'); if (group.systemKey === 'registered') return text('systemRegistered'); return group.joinMode === 'application' ? text('application') : text('inviteOnly') }
function errorMessage(reason: unknown, fallback: string) { return reason instanceof Error && reason.message ? reason.message : fallback }
