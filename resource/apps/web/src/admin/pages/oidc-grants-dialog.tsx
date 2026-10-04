import { useCallback, useEffect, useRef, useState } from 'react'
import type { GooseAdminApi, OIDCAdminGrant, OIDCClient } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Input } from '@gooseforum/ui/components/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@gooseforum/ui/components/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table'
import { Search, Ban, ChevronLeft, ChevronRight } from 'lucide-react'
import { toast } from 'sonner'
import type { IdentityTextKey } from '../identity-settings-i18n'

export function OIDCGrantsDialog({ api, client, text, onClose }: { api: GooseAdminApi; client: OIDCClient; text(k: IdentityTextKey): string; onClose(): void }) {
  const [rows, setRows] = useState<OIDCAdminGrant[]>([])
  const [more, setMore] = useState(false)
  const [cursors, setCursors] = useState([''])
  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState('')
  const [busy, setBusy] = useState(false)
  const [failed, setFailed] = useState(false)
  const [revoking, setRevoking] = useState<string | null>(null)
  const requestId = useRef(0)
  const tr = useRef(text)
  useEffect(() => { tr.current = text }, [text])
  const cursor = cursors[cursors.length - 1]
  const load = useCallback(async () => {
    const id = ++requestId.current
    setBusy(true); setFailed(false)
    try {
      const result = await api.settings.oidcClientGrants(client.clientId, cursor, filter)
      if (id === requestId.current) { setRows(result.items); setMore(result.hasMore) }
    } catch (r) { if (id === requestId.current) { setFailed(true); toast.error(r instanceof Error ? r.message : tr.current('loadFailed')) } }
    finally { if (id === requestId.current) setBusy(false) }
  }, [api, client.clientId, cursor, filter])
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0)
    return () => { window.clearTimeout(timer) }
  }, [load])
  const date = (value: string) => new Date(value).toLocaleString(undefined, { hour12: false })
  return <>
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-3xl">
        <DialogHeader><DialogTitle>{client.name} · {text('grants')}</DialogTitle><DialogDescription>{client.clientId}</DialogDescription></DialogHeader>
        <form className="flex gap-2" onSubmit={(e) => { e.preventDefault(); setFilter(query.trim()); setCursors(['']) }}>
          <Input aria-label={text('userId')} placeholder={text('userId')} inputMode="numeric" value={query} onChange={(e) => setQuery(e.target.value)} />
          <Button type="submit" disabled={busy}><Search />{text('search')}</Button>
        </form>
        {failed ? <Button variant="outline" onClick={() => void load()}>{text('retry')}</Button> : busy ? <p role="status">{text('loading')}</p> : <Table>
          <TableHeader><TableRow><TableHead>{text('userId')}</TableHead><TableHead>{text('scopes')}</TableHead><TableHead>{text('grantedAt')}</TableHead><TableHead>{text('updatedAt')}</TableHead><TableHead /></TableRow></TableHeader>
          <TableBody>{rows.map((row) => <TableRow key={row.userId}><TableCell><span className="block">{row.username || row.userId}</span>{row.username ? <span className="text-xs text-muted-foreground">ID {row.userId}</span> : null}</TableCell><TableCell className="max-w-48 whitespace-normal">{row.scopes.join(', ')}</TableCell><TableCell>{date(row.createdAt)}</TableCell><TableCell>{date(row.updatedAt)}</TableCell><TableCell><Button variant="ghost" size="icon-sm" aria-label={text('revokeGrant')} title={text('revokeGrant')} onClick={() => setRevoking(row.userId)}><Ban /></Button></TableCell></TableRow>)}{!rows.length ? <TableRow><TableCell colSpan={5}>{text('empty')}</TableCell></TableRow> : null}</TableBody>
        </Table>}
        <DialogFooter className="sm:justify-between">
          <Button variant="destructive" disabled={busy || failed} onClick={() => setRevoking('')}>{text('revokeAllGrants')}</Button>
          <div className="flex gap-2"><Button variant="outline" disabled={busy || cursors.length === 1} onClick={() => setCursors(cursors.slice(0, -1))}><ChevronLeft />{text('previous')}</Button><Button variant="outline" disabled={busy || failed || !more} onClick={() => setCursors([...cursors, rows.at(-1)!.userId])}>{text('next')}<ChevronRight /></Button></div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
    <Dialog open={revoking !== null} onOpenChange={(open) => !open && !busy && setRevoking(null)}><DialogContent><DialogHeader><DialogTitle>{text(revoking === '' ? 'revokeAllGrants' : 'revokeGrant')}</DialogTitle><DialogDescription>{client.name}{revoking ? ` · ID ${revoking}` : ''}: {text('revokeGrantHint')}</DialogDescription></DialogHeader><DialogFooter><Button variant="outline" disabled={busy} onClick={() => setRevoking(null)}>{text('cancel')}</Button><Button variant="destructive" disabled={busy} onClick={async () => {
      setBusy(true)
      try { await api.settings.revokeOIDCClientGrant(client.clientId, revoking || undefined); setRevoking(null); await load(); toast.success(text('saved')) }
      catch (r) { toast.error(r instanceof Error ? r.message : text('saveFailed')) }
      finally { setBusy(false) }
    }}>{text('confirm')}</Button></DialogFooter></DialogContent></Dialog>
  </>
}
