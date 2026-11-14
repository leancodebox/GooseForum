import { useLatestRequest } from '@gooseforum/runtime';
import { useEffect, useRef, useState } from 'react'
import { trustLevelName, type GooseAdminApi, type TrustConfig, type TrustLimits } from '@gooseforum/client'
import { trustQuotaFields } from '@gooseforum/client/trust-limits'
import { trustLimitOrderIssue } from '@gooseforum/client/trust-validation'
import { Button } from '@gooseforum/ui/components/button'
import { Field, FieldLabel, FieldLegend, FieldSet } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Switch } from '@gooseforum/ui/components/switch'
import { Checkbox } from '@gooseforum/ui/components/checkbox'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@gooseforum/ui/components/tabs'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Save, RefreshCw } from 'lucide-react'
import { toast } from '@gooseforum/ui/components/sonner'
import type { TrustSettingsText } from '../trust-settings-i18n'
import { errorMessage } from '../error-message'

export function TrustSettingsPage({ api, locale, text }: { api: GooseAdminApi; locale: string; text: TrustSettingsText }) {
  const [form, setForm] = useState<TrustConfig>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [reload, setReload] = useState(0)
  const busy = useRef(false)
  const finiteLimits = useRef(new Map<string, number>())
  const beginRequest = useLatestRequest();
  useEffect(() => {
    const isCurrent = beginRequest()
    void api.trust.settings().then(value => { if (isCurrent()) { finiteLimits.current.clear(); setForm(value) } }).catch(reason => { if (isCurrent()) setError(errorMessage(reason, text('loadFailed'))) }).finally(() => { if (isCurrent()) setLoading(false) })
      }, [api, reload, text, beginRequest])
  async function save(event: React.FormEvent) {
    event.preventDefault()
    if (!form || busy.current) return
    const issue = trustLimitOrderIssue(form.levelLimits)
    if (issue) {
      const lower = trustLevelName(issue.lowerLevel, locale), higher = trustLevelName(issue.higherLevel, locale)
      setError(text('orderInvalid', { higher, field: text(issue.field), lower }))
      return
    }
    busy.current = true; setSaving(true); setError('')
    try { await api.trust.saveSettings(form); toast.success(text('saved')) }
    catch (reason) { setError(errorMessage(reason, text('saveFailed'))) }
    finally { busy.current = false; setSaving(false) }
  }
  function updateLimit(index: number, patch: Partial<TrustLimits>) {
    setForm(current => current && { ...current, levelLimits: current.levelLimits.map((value, i) => i === index ? { ...value, ...patch } : value) })
  }
  return <main className="flex flex-1 flex-col gap-5 px-3 py-3 lg:px-4">
    <form onSubmit={event => void save(event)} className="w-full max-w-6xl space-y-5">
      {error ? <div role="alert" className="flex flex-wrap items-center gap-2 text-sm text-destructive">{error}<Button type="button" size="sm" variant="outline" disabled={saving} onClick={() => { setLoading(true); setError(''); setReload(value => value + 1) }}><RefreshCw />{text('reload')}</Button></div> : null}
      {loading ? <Spinner /> : form ? <fieldset disabled={saving} className="min-w-0 space-y-5">
        <Tabs defaultValue="promotion">
          <TabsList><TabsTrigger value="promotion">{text('promotion')}</TabsTrigger><TabsTrigger value="limits">{text('limits')}</TabsTrigger><TabsTrigger value="cooldown">{text('cooldown')}</TabsTrigger></TabsList>
          <TabsContent value="promotion" className="space-y-5">
            {form.promotionRules.map((rule, index) => <section key={rule.level} className="space-y-3 border-t pt-4">
              <div className="flex items-center gap-3"><h3 className="text-sm font-semibold">{trustLevelName(rule.level, locale)}</h3><Switch aria-label={trustLevelName(rule.level, locale)} checked={rule.enabled} onCheckedChange={enabled => setForm({ ...form, promotionRules: form.promotionRules.map((r, i) => i === index ? { ...r, enabled } : r) })} /></div>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">{([
                ['minAccountAgeHours', 87600],
                ['minVisitDays', 36500],
                ['minTopicsVisited', 20],
              ] as const).map(([key, max]) => <Field key={key} className="min-w-0"><FieldLabel htmlFor={`trust-${rule.level}-${key}`}>{text(key)}</FieldLabel><Input id={`trust-${rule.level}-${key}`} type="number" required min={0} max={max} step={1} value={rule[key]} onChange={event => setForm({ ...form, promotionRules: form.promotionRules.map((r, i) => i === index ? { ...r, [key]: Number(event.target.value) } : r) })} /></Field>)}</div>
            </section>)}
            <div className="flex gap-3 border-t pt-4 text-sm"><span className="font-medium">{trustLevelName(1000, locale)}</span><span className="text-muted-foreground">{text('grantedByAdmin')}</span></div>
          </TabsContent>
          <TabsContent value="limits" className="space-y-5">
            {form.levelLimits.map((limit, index) => <section key={limit.level} className="space-y-3 border-t pt-4"><h3 className="text-sm font-semibold">{trustLevelName(limit.level, locale)}</h3>
              <div className="flex flex-wrap items-center gap-x-8 gap-y-3"><Field orientation="horizontal" className="w-fit max-w-full gap-3 [&>[data-slot=field-label]]:flex-none"><FieldLabel htmlFor={`trust-upload-${limit.level}`}>{text('allowImageUpload')}</FieldLabel><Switch id={`trust-upload-${limit.level}`} checked={limit.allowImageUpload} onCheckedChange={value => updateLimit(index, { allowImageUpload: value })} /></Field>
              <Field orientation="horizontal" className="w-fit max-w-full gap-3 [&>[data-slot=field-label]]:flex-none"><FieldLabel htmlFor={`trust-attachment-${limit.level}`}>{text('allowAttachmentUpload')}</FieldLabel><Switch id={`trust-attachment-${limit.level}`} checked={limit.allowAttachmentUpload ?? false} onCheckedChange={value => updateLimit(index, { allowAttachmentUpload: value })} /></Field>
              <Field orientation="horizontal" className="w-fit max-w-full gap-3 [&>[data-slot=field-label]]:flex-none"><FieldLabel htmlFor={`trust-cover-${limit.level}`}>{text('allowProfileCover')}</FieldLabel><Switch id={`trust-cover-${limit.level}`} checked={limit.allowProfileCover ?? false} onCheckedChange={value => updateLimit(index, { allowProfileCover: value })} /></Field></div>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">{trustQuotaFields.map(key => <FieldSet key={key} className="min-w-0 gap-2" data-quota-field>
                <FieldLegend id={`trust-${limit.level}-${key}-label`} title={key === 'maxDailyTopics' ? text('dailyTopicsHint') : undefined} variant="label" className="mb-2">{text(key)}</FieldLegend>
                <div className="mt-auto grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
                  <Field className="min-w-0">
                    <Input id={`trust-${limit.level}-${key}`} aria-labelledby={`trust-${limit.level}-${key}-label`} type="number" required min={0} max={(key === 'maxDailyTopics' || key === 'maxDailyReplies' || key === 'maxDailyUploads') ? 100000 : 100} step={1} placeholder={text('unlimited')} disabled={saving || limit[key] === null} value={limit[key] ?? ''} onChange={event => updateLimit(index, { [key]: Number(event.target.value) })} />
                  </Field>
                  <label className="flex size-8 shrink-0 cursor-pointer items-center justify-center" title={text('unlimited')}>
                    <Checkbox aria-label={`${text(key)} ${text('unlimited')}`} disabled={saving} checked={limit[key] === null} onCheckedChange={checked => {
                      const cacheKey = `${limit.level}:${key}`
                      if (checked === true) {
                        const value = limit[key]
                        if (value !== null) finiteLimits.current.set(cacheKey, value)
                        updateLimit(index, { [key]: null })
                      } else updateLimit(index, { [key]: finiteLimits.current.get(cacheKey) ?? 0 })
                    }} />
                    <span className="sr-only">{text('unlimited')}</span>
                  </label>
                </div>
              </FieldSet>)}</div>
            </section>)}
          </TabsContent>
          <TabsContent value="cooldown" className="border-t pt-4">
            <div className="grid gap-4 sm:grid-cols-2">
              {(['minPostAccountAgeMinutes', 'minUploadAccountAgeMinutes'] as const).map(key => <Field key={key}><FieldLabel htmlFor={`trust-${key}`}>{text(key)}</FieldLabel><Input id={`trust-${key}`} type="number" min={0} max={5256000} step={1} required value={form.registrationCooldown[key]} onChange={event => setForm({...form,registrationCooldown:{...form.registrationCooldown,[key]:Number(event.target.value)}})} /></Field>)}
            </div>
          </TabsContent>
        </Tabs>
      </fieldset> : null}
      <footer className="flex justify-end border-t pt-4"><Button type="submit" size="sm" disabled={!form || loading || saving}>{saving ? <Spinner /> : <Save />}{text('save')}</Button></footer>
    </form>
  </main>
}
