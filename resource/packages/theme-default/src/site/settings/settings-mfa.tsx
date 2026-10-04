import { useEffect, useState, type FormEvent } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Download, RefreshCw, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useGooseRuntime, GooseLink } from '@gooseforum/runtime'
import { Button } from '@gooseforum/ui/components/button'
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { SettingsSectionHeader } from './settings-section-header'
import { useServerErrorMessage } from '@gooseforum/runtime/i18n/server-error'

export function MFASettings({ showError, onEnabled }: { showError(message: string): void; onEnabled?(enabled: boolean): void }) {
  const { t } = useTranslation('settings')
  const runtime = useGooseRuntime()
  const serverError = useServerErrorMessage()
  const [status, setStatus] = useState<{ enabled: boolean; available: boolean; remainingCodes: number } | null>(null)
  const [setup, setSetup] = useState<{ secret: string; uri: string } | null>(null)
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [pending, setPending] = useState(false)
  const [codes, setCodes] = useState<string[] | null>(null)
  const [statusError, setStatusError] = useState('')
  const [retry, setRetry] = useState(0)
  useEffect(() => {
    let active = true
    if (!runtime.api.users.mfaStatus) return
    void Promise.resolve().then(() => {
      if (!active) return
      setStatusError('')
      return runtime.api.users.mfaStatus()
    }).then((result) => {
      if (!result) return
      if (active) { setStatus(result); onEnabled?.(result.enabled) }
    }).catch((reason) => { if (active) setStatusError(serverError(reason, t('mfa.statusFailed'))) })
    return () => { active = false }
  }, [runtime.api.users, onEnabled, serverError, t, retry])

  async function submit(event: FormEvent, action: 'enable' | 'disable' | 'regenerate' = status?.enabled ? 'regenerate' : 'enable') {
    event.preventDefault()
    if (pending || !password || ((setup || status?.enabled) && !code.trim())) return
    setPending(true)
    try {
      if (!status?.enabled && !setup) {
        const result = await runtime.api.users.mfaBegin(password)
        setSetup(result)
      } else {
        const result = await runtime.api.users.mfaChange(password, code.trim(), action)
        setPassword(''); setCode(''); setSetup(null)
        if (result.recoveryCodes?.length) setCodes(result.recoveryCodes)
        else runtime.redirect('/login')
      }
    } catch (reason) { showError(serverError(reason, t('mfa.failed'))) } finally { setPending(false) }
  }
  function download() {
    if (!codes) return
    const url = URL.createObjectURL(new Blob([codes.join('\n') + '\n'], { type: 'text/plain' }))
    const anchor = document.createElement('a')
    anchor.href = url; anchor.download = 'gooseforum-recovery-codes.txt'; anchor.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }
  if (!status && !runtime.api.users.mfaStatus) return null
  return (
    <section>
      <SettingsSectionHeader icon={ShieldCheck} title={t('mfa.title')} />
      <div className="max-w-3xl p-4 lg:p-5">
        {!status ? statusError ? <Alert><AlertDescription className="flex flex-wrap items-center gap-2"><span className="flex-1">{statusError}</span><Button type="button" variant="outline" size="sm" onClick={() => setRetry((value) => value + 1)}><RefreshCw />{t('mfa.retry')}</Button></AlertDescription></Alert> : <div role="status" className="flex items-center gap-2 text-sm text-muted-foreground"><Spinner role="presentation" aria-hidden="true" />{t('authLogs.loading')}</div> : codes ? <FieldGroup>
          <FieldLabel>{t('mfa.recoveryTitle')}</FieldLabel>
          <p className="text-sm text-muted-foreground">{t('mfa.recoveryNotice')}</p>
          <pre className="overflow-x-auto border-y py-3 text-xs leading-6">{codes.join('\n')}</pre>
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" onClick={download}><Download data-icon="inline-start" />{t('mfa.download')}</Button>
            <Button variant="outline" asChild><GooseLink href="/login">{t('mfa.back')}</GooseLink></Button>
          </div>
        </FieldGroup> : <form onSubmit={(event) => void submit(event)}>
          <FieldGroup>
            <p className="text-sm">{t(status.enabled ? 'mfa.enabled' : 'mfa.disabled')}{status.enabled && ` · ${t('mfa.remaining', { count: status.remainingCodes })}`}</p>
            {!status.available ? <p className="text-sm text-muted-foreground">{t('mfa.unavailable')}</p> : <>
              {!setup && <Field>
                <FieldLabel htmlFor="mfa-password">{t('mfa.password')}</FieldLabel>
                <Input id="mfa-password" type="password" autoComplete="current-password" required value={password} onChange={(event) => setPassword(event.target.value)} />
                <FieldDescription>{t('mfa.oauthPassword')}</FieldDescription>
              </Field>}
              <div className={setup ? 'grid items-start gap-5 sm:grid-cols-[192px_minmax(0,1fr)]' : ''}>
              {setup && <div className="flex flex-col gap-2">
                <QRCodeSVG value={setup.uri} size={192} marginSize={4} role="img" aria-label={t('mfa.qr')} className="size-48 max-w-full bg-white" />
              </div>}
              <FieldGroup className="min-w-0">
              {setup && <Field>
                <FieldLabel>{t('mfa.secret')}</FieldLabel><code className="break-all rounded-md border bg-muted/30 p-3 text-sm leading-6 select-all">{setup.secret}</code>
              </Field>}
              {(setup || status.enabled) && <Field>
                <FieldLabel htmlFor="mfa-settings-code">{t('mfa.code')}</FieldLabel>
                <Input id="mfa-settings-code" autoComplete="one-time-code" required maxLength={64} value={code} onChange={(event) => setCode(event.target.value)} />
              </Field>}
              </FieldGroup>
              </div>
              <p className="text-sm text-muted-foreground">{t('mfa.logoutNotice')}</p>
              <div className="flex flex-wrap gap-2">
                {status.enabled ? <>
                  <Button type="submit" variant="outline" disabled={pending}>{t('mfa.regenerate')}</Button>
                  <Button type="button" variant="outline" disabled={pending || !password || !code} onClick={(event) => void submit(event, 'disable')}>{t('mfa.disable')}</Button>
                </> : <Button type="submit" variant="outline" disabled={pending}>{t(setup ? 'mfa.confirm' : 'mfa.enable')}</Button>}
                {setup && <Button type="button" variant="ghost" disabled={pending} onClick={() => { setSetup(null); setPassword(''); setCode('') }}>{t('mfa.cancel')}</Button>}
              </div>
            </>}
          </FieldGroup>
        </form>}
      </div>
    </section>
  )
}
