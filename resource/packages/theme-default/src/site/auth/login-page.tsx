"use client"

import {
  useCallback,
  useEffect,
  useState,
  type ComponentProps,
  type FormEvent,
} from 'react'
import {
  CircleAlertIcon,
  CircleCheckIcon,
  KeyRoundIcon,
} from 'lucide-react'
import {
  loginWithPassword,
  type LayoutPayload,
  type LoginPageProps,
} from '@gooseforum/client'
import { useTranslation } from 'react-i18next'
import { Alert, AlertDescription } from '@gooseforum/ui/components/alert'
import { Button } from '@gooseforum/ui/components/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@gooseforum/ui/components/card'
import { Checkbox } from '@gooseforum/ui/components/checkbox'
import { Field, FieldGroup, FieldLabel } from '@gooseforum/ui/components/field'
import { Input } from '@gooseforum/ui/components/input'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { Tabs, TabsList, TabsTrigger } from '@gooseforum/ui/components/tabs'
import { useGooseRuntime } from '@gooseforum/runtime'
import { AuthBrand } from './auth-brand'
import { AuthLocaleSwitcher } from './auth-locale-switcher'

type Mode = 'login' | 'register' | 'forgot'

interface LoginPageViewProps {
  layout: LayoutPayload
  page: LoginPageProps
}

export function LoginPageView({ layout, page }: LoginPageViewProps) {
  const runtime = useGooseRuntime()
  const { t } = useTranslation('auth')
  const [mode, setMode] = useState<Mode>(page.initialMode || 'login')
  const [captcha, setCaptcha] = useState({ id: '', image: '' })
  const [captchaLoading, setCaptchaLoading] = useState(false)
  const [pending, setPending] = useState<Mode | null>(null)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [loginForm, setLoginForm] = useState({ username: '', password: '', captcha: '' })
  const [registerForm, setRegisterForm] = useState({
    username: '', email: '', password: '', confirmPassword: '', captcha: '', agree: false,
  })
  const [forgotForm, setForgotForm] = useState({ email: '', captcha: '' })

  const refreshCaptcha = useCallback(async () => {
    setCaptchaLoading(true)
    try {
      const nextCaptcha = await runtime.api.auth.captcha()
      setCaptcha({ id: nextCaptcha.captchaId, image: nextCaptcha.captchaImg })
    } catch (nextError) {
      setError(errorMessage(nextError, t('validation.captchaLoadFailed')))
    } finally {
      setCaptchaLoading(false)
    }
  }, [runtime.api.auth, t])

  useEffect(() => {
    void refreshCaptcha()
  }, [refreshCaptcha])

  function switchMode(nextMode: Mode) {
    setMode(nextMode)
    setError('')
    setNotice('')
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const username = loginForm.username.trim()
    const captchaCode = loginForm.captcha.trim()
    if (!username || !loginForm.password || !captchaCode) {
      setError(t('validation.loginRequired'))
      return
    }

    setPending('login')
    setError('')
    try {
      await loginWithPassword(runtime.api.auth, {
        username,
        password: loginForm.password,
        captchaId: captcha.id,
        captchaCode,
      })
      await runtime.navigate(page.redirectUrl || '/', { replace: true })
    } catch (nextError) {
      setError(errorMessage(nextError, t('validation.loginFailed')))
      setLoginForm((current) => ({ ...current, captcha: '' }))
      void refreshCaptcha()
    } finally {
      setPending(null)
    }
  }

  async function handleRegister(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const username = registerForm.username.trim()
    const email = registerForm.email.trim()
    const captchaCode = registerForm.captcha.trim()
    if (!username || !email || !registerForm.password || !captchaCode) {
      setError(t('validation.registerRequired'))
      return
    }
    if (registerForm.password !== registerForm.confirmPassword) {
      setError(t('validation.passwordMismatch'))
      return
    }
    if (!registerForm.agree) {
      setError(t('validation.termsRequired'))
      return
    }

    setPending('register')
    setError('')
    try {
      const result = await runtime.api.auth.register({
        username,
        email,
        password: registerForm.password,
        captchaId: captcha.id,
        captchaCode,
        locale: runtime.locale,
      })
      runtime.queueFlash(result.message || t('validation.registerSuccess'), 'success')
      await runtime.navigate(page.redirectUrl || '/', { replace: true })
    } catch (nextError) {
      setError(errorMessage(nextError, t('validation.registerFailed')))
      setRegisterForm((current) => ({ ...current, captcha: '' }))
      void refreshCaptcha()
    } finally {
      setPending(null)
    }
  }

  async function handleForgot(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const email = forgotForm.email.trim()
    const captchaCode = forgotForm.captcha.trim()
    if (!email || !captchaCode) {
      setError(t('validation.forgotRequired'))
      return
    }

    setPending('forgot')
    setError('')
    try {
      const result = await runtime.api.auth.forgotPassword(email, captcha.id, captchaCode)
      setNotice(result.message || t('server.passwordResetMailQueued'))
      setForgotForm((current) => ({ ...current, captcha: '' }))
      void refreshCaptcha()
    } catch (nextError) {
      setError(errorMessage(nextError, t('validation.resetEmailFailed')))
      setForgotForm((current) => ({ ...current, captcha: '' }))
      void refreshCaptcha()
    } finally {
      setPending(null)
    }
  }

  const title = mode === 'register'
    ? t('registerTitle')
    : mode === 'forgot'
      ? t('forgotTitle')
      : t('loginTitle')
  const subtitle = mode === 'register'
    ? t('registerSubtitle')
    : mode === 'forgot'
      ? t('forgotSubtitle')
      : t('loginSubtitle')
  const showOAuth = mode !== 'forgot' && page.oauthProviders.length > 0

  return (
    <main className="relative flex min-h-svh items-center justify-center bg-muted px-4 py-16 sm:px-6">
      <AuthLocaleSwitcher />

      <Card className="w-full max-w-md">
        <CardHeader>
          <AuthBrand layout={layout} />
          <CardTitle className="text-2xl">
            <h1>{title}</h1>
          </CardTitle>
          <p className="text-sm leading-6 text-muted-foreground">{subtitle}</p>
        </CardHeader>

        <CardContent className="flex flex-col gap-4">
          {mode !== 'forgot'
            ? (
                <Tabs value={mode} onValueChange={(value) => switchMode(value as Mode)}>
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="login">{t('login')}</TabsTrigger>
                    <TabsTrigger value="register">{t('register')}</TabsTrigger>
                  </TabsList>
                </Tabs>
              )
            : null}

          {error
            ? (
                <Alert variant="destructive">
                  <CircleAlertIcon />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )
            : notice
              ? (
                  <Alert>
                    <CircleCheckIcon />
                    <AlertDescription>{notice}</AlertDescription>
                  </Alert>
                )
              : null}

          {mode === 'login'
            ? (
                <form onSubmit={handleLogin} noValidate>
                  <FieldGroup className="gap-3">
                    <TextField
                      id="login-username"
                      label={t('usernameOrEmail')}
                      autoComplete="username"
                      value={loginForm.username}
                      onChange={(event) => setLoginForm((current) => ({ ...current, username: event.target.value }))}
                    />
                    <TextField
                      id="login-password"
                      label={t('password')}
                      type="password"
                      autoComplete="current-password"
                      value={loginForm.password}
                      onChange={(event) => setLoginForm((current) => ({ ...current, password: event.target.value }))}
                    />
                    <CaptchaField
                      value={loginForm.captcha}
                      onChange={(value) => setLoginForm((current) => ({ ...current, captcha: value }))}
                      image={captcha.image}
                      loading={captchaLoading}
                      label={t('captcha')}
                      imageAlt={t('captchaAlt')}
                      refreshLabel={t('refreshCaptcha')}
                      onRefresh={() => void refreshCaptcha()}
                    />
                    <Button type="button" variant="link" className="self-end px-0" onClick={() => switchMode('forgot')}>
                      {t('forgotPassword')}
                    </Button>
                    <SubmitButton pending={pending === 'login'}>{t('login')}</SubmitButton>
                  </FieldGroup>
                </form>
              )
            : mode === 'register'
              ? (
                  <form onSubmit={handleRegister} noValidate>
                    <FieldGroup className="gap-3">
                      <TextField
                        id="register-username"
                        label={t('username')}
                        autoComplete="username"
                        value={registerForm.username}
                        onChange={(event) => setRegisterForm((current) => ({ ...current, username: event.target.value }))}
                      />
                      <TextField
                        id="register-email"
                        label={t('email')}
                        type="email"
                        autoComplete="email"
                        value={registerForm.email}
                        onChange={(event) => setRegisterForm((current) => ({ ...current, email: event.target.value }))}
                      />
                      <TextField
                        id="register-password"
                        label={t('password')}
                        type="password"
                        autoComplete="new-password"
                        value={registerForm.password}
                        onChange={(event) => setRegisterForm((current) => ({ ...current, password: event.target.value }))}
                      />
                      <TextField
                        id="register-confirm-password"
                        label={t('confirmPassword')}
                        type="password"
                        autoComplete="new-password"
                        value={registerForm.confirmPassword}
                        onChange={(event) => setRegisterForm((current) => ({ ...current, confirmPassword: event.target.value }))}
                      />
                      <CaptchaField
                        value={registerForm.captcha}
                        onChange={(value) => setRegisterForm((current) => ({ ...current, captcha: value }))}
                        image={captcha.image}
                        loading={captchaLoading}
                        label={t('captcha')}
                        imageAlt={t('captchaAlt')}
                        refreshLabel={t('refreshCaptcha')}
                        onRefresh={() => void refreshCaptcha()}
                      />
                      <Field orientation="horizontal">
                        <Checkbox
                          id="register-terms"
                          checked={registerForm.agree}
                          onCheckedChange={(checked) => setRegisterForm((current) => ({ ...current, agree: checked === true }))}
                        />
                        <FieldLabel htmlFor="register-terms">{t('agreeTerms')}</FieldLabel>
                      </Field>
                      <SubmitButton pending={pending === 'register'}>{t('createAccount')}</SubmitButton>
                    </FieldGroup>
                  </form>
                )
              : (
                  <form onSubmit={handleForgot} noValidate>
                    <FieldGroup className="gap-3">
                      <TextField
                        id="forgot-email"
                        label={t('registeredEmail')}
                        type="email"
                        autoComplete="email"
                        value={forgotForm.email}
                        onChange={(event) => setForgotForm((current) => ({ ...current, email: event.target.value }))}
                      />
                      <CaptchaField
                        value={forgotForm.captcha}
                        onChange={(value) => setForgotForm((current) => ({ ...current, captcha: value }))}
                        image={captcha.image}
                        loading={captchaLoading}
                        label={t('captcha')}
                        imageAlt={t('captchaAlt')}
                        refreshLabel={t('refreshCaptcha')}
                        onRefresh={() => void refreshCaptcha()}
                      />
                      <SubmitButton pending={pending === 'forgot'}>{t('sendResetEmail')}</SubmitButton>
                      <Button type="button" variant="link" onClick={() => switchMode('login')}>
                        {t('backToLogin')}
                      </Button>
                    </FieldGroup>
                  </form>
                )}
        </CardContent>

        {showOAuth
          ? (
              <CardFooter className="flex-col items-stretch gap-3">
                <p className="text-center text-xs font-medium text-muted-foreground">{t('continueWith')}</p>
                <div className="grid gap-2 sm:grid-cols-2">
                  {page.oauthProviders.map((provider) => (
                    <Button key={provider.key} asChild variant="outline">
                      <a href={provider.loginUrl}>
                        {provider.key === 'github'
                          ? <GithubMark />
                          : <KeyRoundIcon data-icon="inline-start" />}
                        {provider.displayName}
                      </a>
                    </Button>
                  ))}
                </div>
              </CardFooter>
            )
          : null}
      </Card>
    </main>
  )
}

function TextField({ id, label, ...props }: Omit<ComponentProps<typeof Input>, 'id' | 'className'> & {
  id: string
  label: string
}) {
  return (
    <Field>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <Input {...props} id={id} className="h-10" />
    </Field>
  )
}

function CaptchaField({
  value,
  onChange,
  image,
  loading,
  label,
  imageAlt,
  refreshLabel,
  onRefresh,
}: {
  value: string
  onChange(value: string): void
  image: string
  loading: boolean
  label: string
  imageAlt: string
  refreshLabel: string
  onRefresh(): void
}) {
  return (
    <Field>
      <FieldLabel htmlFor="auth-captcha">{label}</FieldLabel>
      <div className="flex gap-3">
        <Input
          id="auth-captcha"
          className="h-10"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete="off"
        />
        <Button
          type="button"
          variant="outline"
          className="h-10 w-28 overflow-hidden p-0"
          onClick={onRefresh}
          aria-label={refreshLabel}
          title={refreshLabel}
        >
          {loading || !image
            ? <Spinner />
            : <img src={image} alt={imageAlt} className="h-full w-full object-cover" />}
        </Button>
      </div>
    </Field>
  )
}

function SubmitButton({ pending, children }: { pending: boolean, children: string }) {
  return (
    <Button type="submit" size="lg" className="w-full" disabled={pending}>
      {pending ? <Spinner data-icon="inline-start" /> : null}
      {children}
    </Button>
  )
}

function GithubMark() {
  return (
    <svg data-icon="inline-start" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
    </svg>
  )
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error && error.message ? error.message : fallback
}
