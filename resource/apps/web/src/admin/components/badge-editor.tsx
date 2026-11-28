import { useState, type FormEvent } from 'react'
import type { AdminBadge, GooseAdminApi } from '@gooseforum/client'
import { Button } from '@gooseforum/ui/components/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@gooseforum/ui/components/dialog'
import { Field, FieldDescription, FieldLabel } from '@gooseforum/ui/components/field'
import { ForumBadge } from '@gooseforum/ui/components/forum-badge'
import { Input } from '@gooseforum/ui/components/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@gooseforum/ui/components/select'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Switch } from '@gooseforum/ui/components/switch'
import { Textarea } from '@gooseforum/ui/components/textarea'
import { toast } from '@gooseforum/ui/components/sonner'
import { Check, ImagePlus, Link, Upload } from 'lucide-react'
import { errorMessage } from '../error-message'
import { badgeColors, badgeColorValues, badgeLevels, badgeLevelLabel, badgeRuleLabel, builtinBadgeIcons, type BadgeText } from './badge-options'
import type { AssetTextKey } from '../assets-i18n'

type IconMode = 'builtin' | 'upload' | 'link'
const acceptedImageTypes = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']

export function BadgeEditor({ value, api, text, onClose, onSaved }: {
  value: AdminBadge; api: GooseAdminApi; text: BadgeText; onClose(): void; onSaved(): Promise<void>
}) {
  const [form, setForm] = useState<AdminBadge>({ ...value, grantMode: value.type === 'custom' ? 'manual' : value.grantMode })
  const [mode, setMode] = useState<IconMode>(builtinBadgeIcons.some(icon => value.iconUrl === `/static/badges/${icon}.svg`) ? 'builtin' : 'upload')
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [failedImage, setFailedImage] = useState('')
  const [confirmDisabled, setConfirmDisabled] = useState(false)
  const busy = saving || uploading
  const grantedCount = value.grantedCount ?? 0
  const disablingAwardedBadge = value.isEnabled && !form.isEnabled && grantedCount > 0
  const previewBadge = { ...form, name: form.name.trim() || text('previewBadgeName') }

  async function upload(file?: File) {
    if (!file || busy) return
    if (!acceptedImageTypes.includes(file.type)) { setError(text('badgeImageTypeError')); return }
    setUploading(true)
    setError('')
    try {
      const result = await api.pages.uploadImage(file)
      if (!result.url) throw new Error(text('uploadFailed'))
      setForm(current => ({ ...current, iconType: 'asset', iconKey: '', iconUrl: result.url! }))
      setFailedImage('')
    } catch (reason) {
      setError(errorMessage(reason, text('uploadFailed')))
    } finally { setUploading(false) }
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    if (!form.name.trim()) { setError(text('required')); return }
    const iconUrl = form.iconUrl.trim()
    if (!(iconUrl.startsWith('/') && !iconUrl.startsWith('//')) && !/^https?:\/\//i.test(iconUrl)) {
      setError(text('badgeImageUrlError')); return
    }
    if (failedImage === form.iconUrl) { setError(text('badgeImageFailed')); return }
    if (!Number.isSafeInteger(form.sortOrder)) { setError(text('badgeSortError')); return }
    if (disablingAwardedBadge && !confirmDisabled) { setError(text('disableConfirm', { count: grantedCount })); return }
    setSaving(true)
    setError('')
    try {
      await api.assets.saveBadge({ ...form, code: form.code.trim(), name: form.name.trim(), description: form.description.trim(), iconUrl })
      toast.success(text('saved'))
      await onSaved()
      onClose()
    } catch (reason) { setError(errorMessage(reason, text('saveFailed'))) }
    finally { setSaving(false) }
  }

  return <Dialog open onOpenChange={open => { if (!open && !busy) onClose() }}>
    <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden sm:max-w-4xl" showCloseButton={!busy}>
      <DialogHeader className="shrink-0 pr-6">
        <DialogTitle>{value.code ? text('editBadge') : text('newBadge')}</DialogTitle>
        <DialogDescription>{text('badgeEditorHint')}</DialogDescription>
      </DialogHeader>
      <form id="badge-form" onSubmit={submit} className="-m-1 min-h-0 scroll-p-1 overflow-y-auto p-1">
        <div className="grid gap-5 md:grid-cols-[minmax(0,1fr)_15rem]">
          <fieldset disabled={busy} className="min-w-0 space-y-5">
            <Field><FieldLabel htmlFor="badge-name">{text('name')}</FieldLabel><Input id="badge-name" value={form.name} autoFocus maxLength={64} onChange={event => setForm({ ...form, name: event.target.value })} /></Field>
            <Field><FieldLabel htmlFor="badge-description">{text('description')}</FieldLabel><Textarea id="badge-description" value={form.description} maxLength={255} rows={2} onChange={event => setForm({ ...form, description: event.target.value })} /></Field>
            <Field>
              <FieldLabel>{text('badgeIcon')}</FieldLabel>
              <div className="flex flex-wrap gap-1 rounded-lg bg-muted/60 p-1" role="group" aria-label={text('iconSource')}>
                {(['builtin', 'upload', 'link'] as const).map(item => <Button key={item} type="button" size="sm" variant={mode === item ? 'secondary' : 'ghost'} aria-pressed={mode === item} onClick={() => setMode(item)}>
                  {item === 'builtin' ? <ImagePlus data-icon="inline-start" /> : item === 'upload' ? <Upload data-icon="inline-start" /> : <Link data-icon="inline-start" />}
                  {text(item === 'builtin' ? 'builtinIcons' : item === 'upload' ? 'uploadImage' : 'imageLink')}
                </Button>)}
              </div>
              {mode === 'builtin' ? <div className="grid grid-cols-[repeat(auto-fill,3rem)] justify-start gap-2 rounded-lg border p-3" role="group" aria-label={text('builtinIcons')}>
                {builtinBadgeIcons.map(icon => {
                  const url = `/static/badges/${icon}.svg`
                  const label = text(`icon_${icon.replaceAll('-', '_')}` as AssetTextKey)
                  return <Button key={icon} type="button" variant="ghost" aria-label={label} title={label} aria-pressed={form.iconUrl === url} className="relative size-12 min-w-0 p-0 border border-transparent data-[selected=true]:border-primary data-[selected=true]:bg-primary/5" data-selected={form.iconUrl === url} onClick={() => { setForm({ ...form, iconType: 'asset', iconKey: '', iconUrl: url }); setFailedImage('') }}>
                    <img src={url} alt="" className="size-7 object-contain" />{form.iconUrl === url ? <Check className="absolute right-1 top-1 size-3 text-primary" /> : null}
                  </Button>
                })}
              </div> : mode === 'upload' ? <div className="flex flex-col items-center gap-2 rounded-lg border border-dashed p-5 text-center">
                <Button asChild variant="outline"><label className={busy ? 'pointer-events-none' : 'cursor-pointer'}>
                  {uploading ? <Spinner data-icon="inline-start" /> : <Upload data-icon="inline-start" />}{text(form.iconUrl ? 'replaceImage' : 'uploadImage')}
                  <input aria-label={text('uploadImage')} className="sr-only" type="file" accept={acceptedImageTypes.join(',')} disabled={busy} onChange={event => { void upload(event.target.files?.[0]); event.target.value = '' }} />
                </label></Button><p className="text-xs text-muted-foreground">{text('badgeUploadHint')}</p>
              </div> : <Field><FieldLabel htmlFor="badge-image-url">{text('imageLink')}</FieldLabel><Input id="badge-image-url" value={form.iconUrl} placeholder="https://…" onChange={event => { setForm({ ...form, iconType: 'asset', iconKey: '', iconUrl: event.target.value }); setFailedImage('') }} /><FieldDescription>{text('badgeImageUrlHint')}</FieldDescription></Field>}
              {failedImage === form.iconUrl ? <p role="alert" className="text-xs text-destructive">{text('badgeImageFailed')}</p> : null}
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field><FieldLabel htmlFor="badge-level">{text('level')}</FieldLabel><Select value={form.level} disabled={busy} onValueChange={level => setForm({ ...form, level })}><SelectTrigger id="badge-level" className="w-full"><SelectValue /></SelectTrigger><SelectContent>{badgeLevels.map(level => <SelectItem key={level} value={level}>{badgeLevelLabel(level, text)}</SelectItem>)}</SelectContent></Select></Field>
              <Field><FieldLabel htmlFor="badge-color">{text('color')}</FieldLabel><Select value={form.color} disabled={busy} onValueChange={color => setForm({ ...form, color })}><SelectTrigger id="badge-color" className="w-full"><SelectValue /></SelectTrigger><SelectContent>{badgeColors.map(color => <SelectItem key={color} value={color}><span className="inline-block size-3 rounded-full" style={{ backgroundColor: badgeColorValues[color] }} />{text(`color_${color}`)}</SelectItem>)}</SelectContent></Select></Field>
            </div>
            <p className="-mt-2 text-xs text-muted-foreground">{text('badgeColorHint')}</p>
            <div className="space-y-1 rounded-lg bg-muted/50 p-3"><p className="text-sm font-medium">{text('grantMode')} · {text(form.grantMode === 'auto' ? 'auto' : 'manual')}</p><p className="text-xs text-muted-foreground">{badgeRuleLabel(form, text)}</p>{value.type === 'custom' && value.grantMode === 'auto' ? <p className="text-xs text-amber-700 dark:text-amber-400">{text('customAutoCorrection')}</p> : null}</div>
            <Field orientation="horizontal"><div className="flex-1"><FieldLabel htmlFor="badge-enabled">{text('enabled')}</FieldLabel><FieldDescription>{text('badgeEnabledHint')}</FieldDescription></div><Switch id="badge-enabled" checked={form.isEnabled} disabled={busy} onCheckedChange={isEnabled => { setForm({ ...form, isEnabled }); setConfirmDisabled(false) }} /></Field>
            {disablingAwardedBadge ? <label className="flex items-start gap-2 rounded-lg border border-amber-300 bg-amber-50 p-3 text-xs text-amber-950 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100"><input type="checkbox" className="mt-0.5" checked={confirmDisabled} onChange={event => setConfirmDisabled(event.target.checked)} />{text('disableConfirm', { count: grantedCount })}</label> : null}
            <Field orientation="horizontal"><div className="flex-1"><FieldLabel htmlFor="badge-wearable">{text('allowWearing')}</FieldLabel><FieldDescription>{text('badgeWearableHint')}</FieldDescription></div><Switch id="badge-wearable" checked={form.isWearable} disabled={busy} onCheckedChange={isWearable => setForm({ ...form, isWearable })} /></Field>
            <details className="rounded-lg border p-3"><summary className="cursor-pointer text-sm font-medium">{text('advancedSettings')}</summary><div className="mt-3 grid gap-3 sm:grid-cols-2"><Field><FieldLabel htmlFor="badge-code">{text('code')}</FieldLabel><Input id="badge-code" value={form.code || text('autoGeneratedCode')} readOnly className="text-muted-foreground" /></Field><Field><FieldLabel htmlFor="badge-sort">{text('sort')}</FieldLabel><Input id="badge-sort" type="number" step={1} value={Number.isFinite(form.sortOrder) ? form.sortOrder : ''} onChange={event => setForm({ ...form, sortOrder: event.target.valueAsNumber })} /><FieldDescription>{text('badgeSortHint')}</FieldDescription></Field></div></details>
          </fieldset>
          <aside className="min-w-0 space-y-4 rounded-xl border bg-muted/20 p-4 md:sticky md:top-0 md:self-start" aria-label={text('livePreview')}>
            <p className="text-xs font-medium text-muted-foreground">{text('livePreview')}</p>
            <div className="flex flex-col items-center gap-3 rounded-lg border bg-background p-5 text-center">
              <ForumBadge badge={previewBadge} size="lg" onImageError={() => setFailedImage(form.iconUrl)} />
              <div className="min-w-0 max-w-full"><p className="break-words font-semibold">{previewBadge.name}</p><p className="mt-1 text-xs text-muted-foreground">{badgeLevelLabel(form.level, text)}</p></div>
              {form.description ? <p className="w-full break-words text-xs text-muted-foreground">{form.description}</p> : null}
            </div>
            <div><p className="mb-2 text-xs text-muted-foreground">{text('wearingPreview')}</p><div className="flex items-center gap-3 rounded-lg border bg-background p-3"><div className="relative size-10 shrink-0 rounded-full bg-muted"><img src="/static/pic/default-avatar.webp" alt="" className="size-full rounded-full object-cover" />{form.isWearable ? <ForumBadge badge={previewBadge} size="sm" variant="worn" className="absolute -bottom-1 -right-1 size-5" /> : null}</div><div className="min-w-0"><p className="text-sm font-medium">{text('previewMemberName')}</p><p className="text-xs text-muted-foreground">{text(form.isWearable ? 'wearingPreviewHint' : 'wearingDisabledHint')}</p></div></div></div>
            {!form.isEnabled ? <p className="text-xs text-muted-foreground">{text('disabledPreviewHint')}</p> : null}
          </aside>
        </div>
      </form>
      {error ? <p role="alert" className="shrink-0 text-sm text-destructive">{error}</p> : null}
      <DialogFooter className="shrink-0"><Button variant="outline" disabled={busy} onClick={onClose}>{text('cancel')}</Button><Button form="badge-form" type="submit" disabled={busy}>{busy ? <Spinner data-icon="inline-start" /> : null}{text(saving ? 'saving' : 'save')}</Button></DialogFooter>
    </DialogContent>
  </Dialog>
}
