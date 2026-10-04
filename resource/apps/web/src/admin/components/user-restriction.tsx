import { useEffect, useState } from 'react'
import type { GooseAdminApi, UserRestrictionHistory } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { DatePicker } from '@gooseforum/ui/components/date-picker'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@gooseforum/ui/lib/utils'
import type { UserTextKey } from '../users-i18n'
import type { RestrictionForm } from './user-restriction-form'

type Text = (key: UserTextKey) => string

export function RestrictionFields({ value, initialStatus, onChange, text, locale = 'en', reasonClassName }: { value: RestrictionForm; initialStatus: RestrictionForm['status']; onChange(value: RestrictionForm): void; text: Text; locale?: string; reasonClassName?: string }) {
  return <>
    <Field><FieldLabel htmlFor="restriction-status">{text('status')}</FieldLabel><Select value={value.status} onValueChange={(status) => onChange({ ...value, status: status as RestrictionForm['status'] })}><SelectTrigger id="restriction-status" className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="normal">{text('enabled')}</SelectItem><SelectItem value="suspended">{text('suspended')}</SelectItem><SelectItem value="banned">{text('banned')}</SelectItem></SelectContent></Select></Field>
    {value.status !== 'normal' ? <Field><FieldLabel htmlFor="restriction-until">{text('restrictionUntil')}</FieldLabel><DatePicker id="restriction-until" withTime locale={locale} value={value.until} onChange={(until) => onChange({ ...value, until })} placeholder={text('restrictionNoEnd')} clearLabel={text('dateClear')} timeLabel={text('dateTime')} doneLabel={text('dateDone')} /><span className="text-xs text-muted-foreground">{text('restrictionPermanent')}</span></Field> : null}
    {value.status !== 'normal' || initialStatus !== 'normal' ? <Field className={reasonClassName}><FieldLabel htmlFor="restriction-reason">{text(value.status === 'normal' ? 'restrictionLiftReason' : 'restrictionReason')}</FieldLabel><Input id="restriction-reason" maxLength={500} required value={value.reason} onChange={(event) => onChange({ ...value, reason: event.target.value })} /></Field> : null}
  </>
}

export function RestrictionHistory({ userId, api, text, className }: { userId?: number; api: GooseAdminApi; text: Text; className?: string }) {
  return userId ? <UserHistory key={userId} userId={userId} api={api} text={text} className={className} /> : null
}

function UserHistory({ userId, api, text, className }: { userId: number; api: GooseAdminApi; text: Text; className?: string }) {
  const [page, setPage] = useState(1)
  const [result, setResult] = useState<{ page: number; records: UserRestrictionHistory[]; total: number; error: string }>({page:0,records:[],total:0,error:''})
  const loading = result.page !== page
  const { records, total, error } = result
  useEffect(() => {
    let active = true
    void api.users.restrictionHistory({ userId, page, pageSize: 10 }).then((data) => { if (active) setResult({page,records:data.list || [],total:data.total,error:''}) }).catch((reason: unknown) => { if (active) setResult({page,records:[],total:0,error:reason instanceof Error ? reason.message : text('loadFailed')}) })
    return () => { active = false }
  }, [api, page, userId, text])
  return <section className={cn("mt-4 border-t pt-3", className)}><div className="mb-2 flex items-center justify-between gap-3"><h3 className="text-sm font-semibold">{text('restrictionHistory')}</h3><div className="flex items-center gap-2"><Button type="button" variant="outline" size="icon-sm" title={text('previous')} disabled={loading || page <= 1} onClick={() => setPage(page - 1)}><ChevronLeft /></Button><span className="text-xs">{page}/{Math.max(1, Math.ceil(total / 10))}</span><Button type="button" variant="outline" size="icon-sm" title={text('next')} disabled={loading || page * 10 >= total} onClick={() => setPage(page + 1)}><ChevronRight /></Button></div></div>{error ? <p role="alert" className="text-sm">{error}</p> : loading ? <p className="text-xs text-muted-foreground">{text('loading')}</p> : !records.length ? <p className="text-xs text-muted-foreground">{text('restrictionHistoryEmpty')}</p> : <ol className="divide-y">{records.map((entry) => <li key={entry.id} className="py-2 text-xs"><div className="flex flex-wrap justify-between gap-2"><span>{entry.status === 'banned' ? text('banned') : entry.status === 'suspended' ? text('suspended') : text('enabled')} · #{entry.actorId}</span><time>{new Date(entry.createdAt).toLocaleString()}</time></div><p className="mt-1 break-words">{entry.reason}</p>{entry.until ? <p>{text('restrictionUntil')}: {new Date(entry.until).toLocaleString()}</p> : null}</li>)}</ol>}</section>
}
