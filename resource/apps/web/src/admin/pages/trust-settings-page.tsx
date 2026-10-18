import { useEffect, useRef, useState } from 'react'
import { trustLevelName, type GooseAdminApi, type TrustConfig, type TrustLimits } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Field, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Switch } from '@gooseforum/ui/components/switch'
import { Checkbox } from '@gooseforum/ui/components/checkbox'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@gooseforum/ui/components/tabs'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Save, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'

export function TrustSettingsPage({ api, locale }: { api: GooseAdminApi; locale: string }) {
  const zh = locale.startsWith('zh')
  const [form, setForm] = useState<TrustConfig>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [reload, setReload] = useState(0)
  const busy = useRef(false)
  useEffect(() => {
    let active = true
    void api.trust.settings().then(value => { if (active) setForm(value) }).catch(reason => { if (active) setError(String(reason)) }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [api, reload])
  async function save(event: React.FormEvent) {
    event.preventDefault()
    if (!form || busy.current) return
    busy.current = true; setSaving(true); setError('')
    try { await api.trust.saveSettings(form); toast.success(zh ? '已保存' : 'Saved') }
    catch (reason) { setError(reason instanceof Error ? reason.message : String(reason)) }
    finally { busy.current = false; setSaving(false) }
  }
  function updateLimit(index: number, patch: Partial<TrustLimits>) {
    setForm(current => current && { ...current, levelLimits: current.levelLimits.map((value, i) => i === index ? { ...value, ...patch } : value) })
  }
  return <main className="flex flex-1 flex-col gap-5 px-3 py-3 lg:px-4">
    <form onSubmit={event => void save(event)} className="w-full max-w-6xl space-y-5">
      {error ? <div role="alert" className="flex flex-wrap items-center gap-2 text-sm text-destructive">{error}<Button type="button" size="sm" variant="outline" disabled={saving} onClick={() => { setLoading(true); setError(''); setReload(value => value + 1) }}><RefreshCw />{zh ? '重新加载' : 'Reload'}</Button></div> : null}
      {loading ? <Spinner /> : form ? <fieldset disabled={saving} className="min-w-0 space-y-5">
        <Tabs defaultValue="promotion">
          <TabsList><TabsTrigger value="promotion">{zh ? '晋级条件' : 'Promotion'}</TabsTrigger><TabsTrigger value="limits">{zh ? '等级限制' : 'Limits'}</TabsTrigger></TabsList>
          <TabsContent value="promotion" className="space-y-5">
            <Field orientation="horizontal" className="w-fit gap-3 [&>[data-slot=field-label]]:flex-none"><FieldLabel htmlFor="trust-auto">{zh ? '自动晋级' : 'Automatic promotion'}</FieldLabel><Switch id="trust-auto" checked={form.autoPromotionEnabled} onCheckedChange={value => setForm({ ...form, autoPromotionEnabled: value })} /></Field>
            {form.promotionRules.map((rule, index) => <section key={rule.level} className="space-y-3 border-t pt-4">
              <div className="flex items-center gap-3"><h3 className="text-sm font-semibold">{trustLevelName(rule.level, locale)}</h3><Switch aria-label={trustLevelName(rule.level, locale)} checked={rule.enabled} onCheckedChange={enabled => setForm({ ...form, promotionRules: form.promotionRules.map((r, i) => i === index ? { ...r, enabled } : r) })} /></div>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">{([
                ['minAccountAgeHours', zh ? '注册时长（小时）' : 'Account age (hours)', 87600],
                ['minVisitDays', zh ? '有效访问日（UTC）' : 'Visit days (UTC)', 36500],
                ['minTopicsVisited', zh ? '不同主题数（最多 20）' : 'Distinct topics (max 20)', 20],
              ] as const).map(([key, label, max]) => <Field key={key} className="min-w-0"><FieldLabel htmlFor={`trust-${rule.level}-${key}`}>{label}</FieldLabel><Input id={`trust-${rule.level}-${key}`} type="number" required min={0} max={max} step={1} value={rule[key]} onChange={event => setForm({ ...form, promotionRules: form.promotionRules.map((r, i) => i === index ? { ...r, [key]: Number(event.target.value) } : r) })} /></Field>)}</div>
            </section>)}
            <div className="flex gap-3 border-t pt-4 text-sm"><span className="font-medium">{trustLevelName(1000, locale)}</span><span className="text-muted-foreground">{zh ? '管理员授予' : 'Granted by an administrator'}</span></div>
          </TabsContent>
          <TabsContent value="limits" className="space-y-5">
            {form.levelLimits.map((limit, index) => <section key={limit.level} className="space-y-3 border-t pt-4"><h3 className="text-sm font-semibold">{trustLevelName(limit.level, locale)}</h3>
              <Field orientation="horizontal" className="w-fit gap-3 [&>[data-slot=field-label]]:flex-none"><FieldLabel htmlFor={`trust-upload-${limit.level}`}>{zh ? '允许上传图片' : 'Allow image uploads'}</FieldLabel><Switch id={`trust-upload-${limit.level}`} checked={limit.allowImageUpload} onCheckedChange={value => updateLimit(index, { allowImageUpload: value })} /></Field>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">{([
                ['maxDailyTopics', zh ? '每日新建主题数（含草稿）' : 'Daily new topics (including drafts)'],
                ['maxDailyReplies', zh ? '每日新建回复数' : 'Daily new replies'],
                ['maxDailyUploads', zh ? '每日上传次数' : 'Daily uploads'],
                ['maxImagesPerPost', zh ? '每篇图片数' : 'Images per post'],
                ['maxExternalLinksPerPost', zh ? '每篇外链数' : 'External links per post'],
                ['maxMentionsPerPost', zh ? '每篇提及人数' : 'Mentioned users per post'],
              ] as const).map(([key, label]) => <Field key={key}><FieldLabel htmlFor={`trust-${limit.level}-${key}`} className="lg:min-h-10 lg:items-start">{label}</FieldLabel><Input id={`trust-${limit.level}-${key}`} type="number" required min={0} max={(key === 'maxDailyTopics' || key === 'maxDailyReplies' || key === 'maxDailyUploads') ? 100000 : 100} step={1} disabled={saving || limit[key] === null} value={limit[key] ?? ''} onChange={event => updateLimit(index, { [key]: Number(event.target.value) })} /><label className="flex items-center gap-2 text-xs text-muted-foreground"><Checkbox checked={limit[key] === null} onCheckedChange={value => updateLimit(index, { [key]: value ? null : 0 })} />{zh ? '不限' : 'Unlimited'}</label></Field>)}</div>
              <div className="grid gap-4 sm:grid-cols-2">
                {([['minPostAccountAgeMinutes', zh ? '发帖所需注册时长（分钟）' : 'Account age for posting (minutes)'], ['minUploadAccountAgeMinutes', zh ? '上传所需注册时长（分钟）' : 'Account age for uploads (minutes)']] as const).map(([key,label]) => <Field key={key}><FieldLabel htmlFor={`trust-${limit.level}-${key}`}>{label}</FieldLabel><Input id={`trust-${limit.level}-${key}`} type="number" min={0} max={5256000} step={1} required value={limit[key]} onChange={event => updateLimit(index,{[key]:Number(event.target.value)})} /></Field>)}
              </div>
            </section>)}
          </TabsContent>
        </Tabs>
      </fieldset> : null}
      <footer className="flex justify-end border-t pt-4"><Button type="submit" size="sm" disabled={!form || loading || saving}>{saving ? <Spinner /> : <Save />}{zh ? '保存' : 'Save'}</Button></footer>
    </form>
  </main>
}
