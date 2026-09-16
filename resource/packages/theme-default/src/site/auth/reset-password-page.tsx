"use client"

import { useState, type FormEvent } from 'react'
import { CheckIcon, CircleAlertIcon, CircleCheckIcon } from 'lucide-react'
import type { LayoutPayload, ResetPasswordPageProps } from '@gooseforum/client'
import { useTranslation } from 'react-i18next'
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert'
import { Button } from '@gooseforum/ui/components/button'
import { Card, CardContent, CardHeader, CardTitle } from '@gooseforum/ui/components/card'
import { Field, FieldGroup, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { GooseLink, useGooseRuntime } from '@gooseforum/runtime'
import { useServerErrorMessage } from '@gooseforum/runtime/i18n/server-error'
import { AuthBrand } from './auth-brand'

export function ResetPasswordPageView({
  layout,
  page,
}: {
  layout: LayoutPayload
  page: ResetPasswordPageProps
}) {
  const { t } = useTranslation('auth')
  const runtime = useGooseRuntime()
  const serverError = useServerErrorMessage()
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSuccess('')
    if (!page.token) {
      setError(t('resetMissingToken'))
      return
    }
    if (password.length < 6) {
      setError(t('passwordMinLength'))
      return
    }
    if (password !== confirmPassword) {
      setError(t('validation.passwordMismatch'))
      return
    }

    setPending(true)
    try {
      const result = await runtime.api.auth.resetPassword(page.token, password)
      setSuccess(result.message || t('server.passwordResetSuccess'))
      setPassword('')
      setConfirmPassword('')
    } catch (nextError) {
      setError(serverError(nextError, t('server.passwordResetFailed')))
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center bg-muted px-4 py-12 sm:px-6">
      <Card className="w-full max-w-3xl md:grid md:grid-cols-2">
        <section className="flex flex-col justify-center">
          <CardHeader>
            <AuthBrand layout={layout} />
            <CardTitle className="text-2xl"><h1>{t('resetPasswordTitle')}</h1></CardTitle>
            <p className="text-sm leading-6 text-muted-foreground">{t('resetPasswordSubtitle')}</p>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {!page.token
              ? (
                  <Alert variant="destructive">
                    <CircleAlertIcon />
                    <AlertDescription>{t('resetMissingToken')}</AlertDescription>
                  </Alert>
                )
              : error
                ? (
                    <Alert variant="destructive">
                      <CircleAlertIcon />
                      <AlertDescription>{error}</AlertDescription>
                    </Alert>
                  )
                : success
                  ? (
                      <Alert>
                        <CircleCheckIcon />
                        <AlertDescription>{success}</AlertDescription>
                      </Alert>
                    )
                  : null}

            <form onSubmit={submit} noValidate>
              <FieldGroup className="gap-3">
                <Field>
                  <FieldLabel htmlFor="reset-password">{t('newPassword')}</FieldLabel>
                  <Input
                    id="reset-password"
                    type="password"
                    className="h-10"
                    autoComplete="new-password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="reset-confirm-password">{t('confirmPassword')}</FieldLabel>
                  <Input
                    id="reset-confirm-password"
                    type="password"
                    className="h-10"
                    autoComplete="new-password"
                    value={confirmPassword}
                    onChange={(event) => setConfirmPassword(event.target.value)}
                  />
                </Field>
                <Button
                  type="submit"
                  size="lg"
                  className="w-full"
                  disabled={!page.token || !password || !confirmPassword || pending}
                >
                  {pending ? <Spinner data-icon="inline-start" /> : null}
                  {t('saveNewPassword')}
                </Button>
                <Button asChild variant="link" size="lg">
                  <GooseLink href="/login">{t('backToLogin')}</GooseLink>
                </Button>
              </FieldGroup>
            </form>
          </CardContent>
        </section>

        <aside className="flex flex-col justify-center border-t bg-muted/50 px-6 py-8 md:border-l md:border-t-0">
          <h2 className="text-lg font-semibold">{t('passwordAdviceTitle')}</h2>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">{t('passwordAdviceDescription')}</p>
          <ul className="mt-6 flex flex-col gap-3 text-sm">
            <Advice>{t('passwordAdvice.length')}</Advice>
            <Advice>{t('passwordAdvice.unique')}</Advice>
            <Advice>{t('passwordAdvice.loginAfterReset')}</Advice>
          </ul>
        </aside>
      </Card>
    </main>
  )
}

function Advice({ children }: { children: string }) {
  return (
    <li className="flex items-center gap-2">
      <CheckIcon className="text-primary" aria-hidden="true" />
      {children}
    </li>
  )
}
