import { useMemo, useState, type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { GooseSiteApi, LayoutPayload, LoginPageProps } from '@gooseforum/client'
import { createGooseI18n, GooseI18nProvider } from '../src/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '../src/runtime'
import { LoginPageView } from '../src/site/auth/login-page'
import type { Locale } from '../src/i18n/auth'

afterEach(cleanup)

const layout = {
  site: {
    name: 'GooseForum',
    description: '',
    logo: '',
    favicon: '',
    brandType: 'default',
    brandText: '',
    brandImage: '',
  },
} as LayoutPayload

const loginPage: LoginPageProps = {
  initialMode: 'login',
  redirectUrl: '/',
  oauthProviders: [],
}

function createAuthApi() {
  return {
    captcha: vi.fn().mockResolvedValue({
      captchaId: 'captcha-id',
      captchaImg: 'data:image/png;base64,dGVzdA==',
    }),
    loginPublicKey: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    forgotPassword: vi.fn(),
    resetPassword: vi.fn(),
    logout: vi.fn(),
  }
}

function renderLogin(options: { locale?: Locale, page?: LoginPageProps } = {}) {
  const auth = createAuthApi()
  const user = userEvent.setup()

  function Harness({ children }: { children: ReactNode }) {
    const [locale, setLocale] = useState<Locale>(options.locale || 'zh')
    const runtime = useMemo<GooseRuntime>(() => ({
      api: { auth } as unknown as GooseSiteApi,
      currentUrl: '/login',
      isNavigating: false,
      theme: 'gf-light',
      locale,
      navigate: vi.fn(),
      queueFlash: vi.fn(),
      redirect: vi.fn(),
      refresh: vi.fn(),
      setLocale,
      toggleTheme: vi.fn(),
    }), [locale])

    return (
      <GooseI18nProvider locale={locale}>
        <GooseRuntimeProvider runtime={runtime}>{children}</GooseRuntimeProvider>
      </GooseI18nProvider>
    )
  }

  render(
    <Harness>
      <LoginPageView layout={layout} page={options.page || loginPage} />
    </Harness>,
  )
  return { auth, user }
}

describe('GooseForum i18n', () => {
  it('keeps instances isolated for SSR hosts', async () => {
    const chinese = createGooseI18n('zh')
    const english = createGooseI18n('en')

    expect(chinese.t('login')).toBe('登录')
    expect(english.t('login')).toBe('Log in')

    await english.changeLanguage('ja')
    expect(english.t('login')).toBe('ログイン')
    expect(chinese.t('login')).toBe('登录')
  })
})

describe('LoginPageView', () => {
  it('loads and refreshes the captcha through the runtime API', async () => {
    const { auth, user } = renderLogin()

    expect(await screen.findByRole('img', { name: '验证码' })).toBeTruthy()
    expect(auth.captcha).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: '刷新验证码' }))
    await waitFor(() => expect(auth.captcha).toHaveBeenCalledTimes(2))
  })

  it('validates an empty login without submitting credentials', async () => {
    const { auth, user } = renderLogin()

    await user.click(screen.getByRole('button', { name: '登录' }))

    expect(await screen.findByText('请填写账号、密码和验证码')).toBeTruthy()
    expect(auth.loginPublicKey).not.toHaveBeenCalled()
    expect(auth.login).not.toHaveBeenCalled()
  })

  it('covers register and forgot-password modes without calling invalid APIs', async () => {
    const { auth, user } = renderLogin()

    await user.click(screen.getByRole('tab', { name: '注册' }))
    expect(screen.getByRole('heading', { name: '创建新账号' })).toBeTruthy()
    expect(screen.getByLabelText('我已阅读并同意服务条款和隐私政策')).toBeTruthy()
    await user.click(screen.getByRole('button', { name: '创建账号' }))
    expect(await screen.findByText('请完整填写注册信息')).toBeTruthy()
    expect(auth.register).not.toHaveBeenCalled()

    await user.click(screen.getByRole('tab', { name: '登录' }))
    await user.click(screen.getByRole('button', { name: '忘记密码？' }))
    expect(screen.getByRole('heading', { name: '重置密码' })).toBeTruthy()
    await user.click(screen.getByRole('button', { name: '发送重置邮件' }))
    expect(await screen.findByText('请填写邮箱和验证码')).toBeTruthy()
    expect(auth.forgotPassword).not.toHaveBeenCalled()
  })

  it('updates translations when the host changes locale', async () => {
    const { user } = renderLogin()

    await user.click(screen.getByRole('tab', { name: 'EN' }))

    expect(await screen.findByRole('heading', { name: 'Log in to your account' })).toBeTruthy()
    expect(screen.getByLabelText('Username or email')).toBeTruthy()
  })

  it('honors the initial page mode from the payload', async () => {
    renderLogin({ page: { ...loginPage, initialMode: 'forgot' } })

    expect(await screen.findByRole('heading', { name: '重置密码' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '返回登录' })).toBeTruthy()
  })
})
