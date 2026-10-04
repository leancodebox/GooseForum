import { useEffect, useState } from 'react'
import type { GooseAdminApi } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { ShieldOff, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import type { UserTextKey } from '../users-i18n'

export function UserMFAReset({ userId, username, api, text, disabled }: { userId: number; username: string; api: GooseAdminApi; text(key: UserTextKey): string; disabled: boolean }) {
  const [enabled, setEnabled] = useState<boolean | null>(null)
  const [error, setError] = useState('')
  const [retry, setRetry] = useState(0)
  const [open, setOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [pending, setPending] = useState(false)
  useEffect(() => {
    let active = true
    void api.users.mfaStatus(userId).then(status => { if (active) { setEnabled(status.enabled); setError('') } }).catch(() => { if (active) setError(text('loadFailed')) })
    return () => { active = false }
  }, [api, userId, text, retry])
  async function reset() {
    if (pending || disabled || !reason.trim()) return
    setPending(true)
    try {
      await api.users.resetMFA(userId, reason.trim())
      setEnabled(false); setOpen(false); setReason('')
      toast.success(text('mfaResetSaved'))
    } catch { toast.error(text('loadFailed')) } finally { setPending(false) }
  }
  return <section className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
    <div className="text-sm"><span className="font-medium">{text('mfaTitle')}</span><span className="ml-3 text-xs text-muted-foreground">{error || (enabled === null ? text('loading') : enabled ? text('enabled') : text('mfaNotEnabled'))}</span></div>
    {error ? <Button type="button" size="sm" variant="outline" onClick={() => setRetry(value => value + 1)}><RefreshCw />{text('retry')}</Button> : <Button type="button" size="sm" variant="outline" disabled={disabled || enabled !== true} onClick={() => { setReason(''); setOpen(true) }}><ShieldOff />{text('mfaReset')}</Button>}
    <Dialog open={open} onOpenChange={value => { if (!pending) { setOpen(value); if (!value) setReason('') } }}>
      <DialogContent><DialogHeader><DialogTitle>{text('mfaReset')} · {username}</DialogTitle><DialogDescription>{text('mfaResetHint')}</DialogDescription></DialogHeader>
        <Field><FieldLabel htmlFor="admin-mfa-reason">{text('mfaResetReason')}</FieldLabel><Input id="admin-mfa-reason" maxLength={60} value={reason} onChange={event => setReason(event.target.value)} /></Field>
        <DialogFooter><Button type="button" variant="outline" disabled={pending} onClick={() => setOpen(false)}>{text('cancel')}</Button><Button type="button" variant="destructive" disabled={pending || disabled || !reason.trim()} onClick={() => void reset()}>{text('mfaReset')}</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
}
