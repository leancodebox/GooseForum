"use client"

import { useEffect, useState } from 'react'
import { ArrowRightIcon, CheckIcon, KeyRoundIcon, XIcon } from 'lucide-react'
import type { LayoutPayload, OIDCConsentDetails, OIDCConsentPageProps } from '@gooseforum/client'
import { useTranslation } from 'react-i18next'
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert'
import { Button } from '@gooseforum/ui/components/button'
import { Card, CardContent, CardHeader, CardTitle } from '@gooseforum/ui/components/card'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { GooseLink, useGooseRuntime } from '@gooseforum/runtime'
import { useServerErrorMessage } from '@gooseforum/runtime/i18n/server-error'
import { AuthBrand } from './auth-brand'
import { AuthLocaleSwitcher } from './auth-locale-switcher'

type Decision = 'approve' | 'deny'

export function OIDCConsentPageView({
  layout,
  page,
}: {
  layout: LayoutPayload
  page: OIDCConsentPageProps
}) {
  const { t } = useTranslation('oidcConsent')
  const runtime = useGooseRuntime()
  const serverError = useServerErrorMessage()
  const [details, setDetails] = useState<OIDCConsentDetails>()
  const [error, setError] = useState('')
  const [decision, setDecision] = useState<Decision>()
  const siteName = layout.site.name

  useEffect(() => {
    let active = true
    if (!page.interaction) {
      setError(t('expired'))
      return () => { active = false }
    }

    void runtime.api.oidc.consentDetails(page.interaction)
      .then((result) => {
        if (active) setDetails(result)
      })
      .catch((reason: unknown) => {
        if (active) setError(serverError(reason, t('loadFailed')))
      })
    return () => { active = false }
  }, [page.interaction, runtime.api.oidc, serverError, t])

  async function decide(nextDecision: Decision) {
    if (decision) return
    setDecision(nextDecision)
    setError('')
    try {
      const result = await runtime.api.oidc.consentDecision(page.interaction, nextDecision)
      if (!result.redirect_url) throw new Error(t('decisionFailed'))
      runtime.redirect(result.redirect_url)
    } catch (reason) {
      setError(serverError(reason, t('decisionFailed')))
      setDecision(undefined)
    }
  }

  return (
    <main className="relative flex min-h-svh items-center justify-center bg-muted px-4 py-16 sm:px-6">
      <AuthLocaleSwitcher />
      <Card className="w-full max-w-lg">
        <CardHeader>
          <AuthBrand layout={layout} />
          <CardTitle className="text-2xl"><h1>{t('title', { site: siteName })}</h1></CardTitle>
          <p className="text-sm leading-6 text-muted-foreground">{t('subtitle')}</p>
        </CardHeader>
        <CardContent>
          {!details && !error
            ? (
                <div className="grid min-h-56 place-items-center" role="status">
                  <div className="flex flex-col items-center gap-3 text-sm text-muted-foreground">
                    <Spinner className="size-6 text-primary" aria-hidden="true" />
                    {t('verifying')}
                  </div>
                </div>
              )
            : error && !details
              ? (
                  <div className="flex flex-col gap-5">
                    <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>
                    <Button asChild variant="outline">
                      <GooseLink href="/">{t('back', { site: siteName })}</GooseLink>
                    </Button>
                  </div>
                )
              : details
                ? (
                    <div className="flex flex-col gap-5">
                      {error ? <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert> : null}

                      <div className="flex items-center gap-3 rounded-xl border bg-muted/50 p-4">
                        <span className="grid size-10 shrink-0 place-items-center rounded-lg border bg-background">
                          <KeyRoundIcon aria-hidden="true" />
                        </span>
                        <div className="min-w-0 flex-1">
                          <p className="break-words text-sm font-semibold">{details.client.name}</p>
                          <p className="mt-1 text-xs leading-5 text-muted-foreground">{t('accessAccount', { site: siteName })}</p>
                        </div>
                        <ArrowRightIcon className="shrink-0 text-muted-foreground" aria-hidden="true" />
                      </div>

                      <section>
                        <h2 className="text-sm font-semibold">{t('permissions')}</h2>
                        <ul className="mt-3 divide-y overflow-hidden rounded-xl border">
                          {details.scopes.map((scope) => (
                            <li key={scope} className="flex items-center gap-3 px-4 py-3 text-sm">
                              <CheckIcon className="shrink-0 text-primary" aria-hidden="true" />
                              <span>{t(`scopes.${scope}`, { defaultValue: scope })}</span>
                            </li>
                          ))}
                        </ul>
                      </section>

                      <p className="rounded-lg bg-muted px-3 py-2.5 text-xs leading-5 text-muted-foreground">
                        {t('clientId')} <span className="ml-1 break-all font-mono">{details.client.id}</span>
                      </p>
                      <p className="text-xs leading-5 text-muted-foreground">{t('trust')}</p>

                      <div className="flex flex-wrap gap-3">
                        <Button variant="outline" size="lg" className="flex-1" disabled={Boolean(decision)} onClick={() => void decide('deny')}>
                          {decision === 'deny' ? <Spinner data-icon="inline-start" /> : <XIcon data-icon="inline-start" />}
                          {decision === 'deny' ? t('loading') : t('deny')}
                        </Button>
                        <Button size="lg" className="flex-1" disabled={Boolean(decision)} onClick={() => void decide('approve')}>
                          {decision === 'approve' ? <Spinner data-icon="inline-start" /> : <CheckIcon data-icon="inline-start" />}
                          {decision === 'approve' ? t('loading') : t('approve')}
                        </Button>
                      </div>
                    </div>
                  )
                : null}
        </CardContent>
      </Card>
    </main>
  )
}
