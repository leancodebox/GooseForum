import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { GooseSiteApi, LayoutPayload } from '@gooseforum/client'
import { GooseI18nProvider } from '../src/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '../src/runtime'
import { OIDCConsentPageView } from '../src/site/auth/oidc-consent-page'

afterEach(cleanup)

const layout = {
  site: {
    name: 'GooseForum', description: '', logo: '', favicon: '',
    brandType: 'default', brandText: '', brandImage: '',
  },
} as LayoutPayload

function renderConsent(options: { interaction?: string, detailsError?: Error } = {}) {
  const details = options.detailsError
    ? vi.fn().mockRejectedValue(options.detailsError)
    : vi.fn().mockResolvedValue({
        client: { id: 'wiki-client', name: 'Goose Wiki', public: true },
        scopes: ['openid', 'profile', 'custom_scope'],
        expires_at: '2026-09-14T12:00:00Z',
      })
  const decision = vi.fn().mockResolvedValue({ redirect_url: 'https://client.example/callback' })
  const redirect = vi.fn()
  const runtime: GooseRuntime = {
    api: { oidc: { consentDetails: details, consentDecision: decision } } as unknown as GooseSiteApi,
    currentUrl: '/oauth2/consent', isNavigating: false, locale: 'zh', theme: 'gf-light',
    navigate: vi.fn(), queueFlash: vi.fn(), redirect, refresh: vi.fn(), setLocale: vi.fn(), toggleTheme: vi.fn(),
  }

  render(
    <GooseI18nProvider locale="zh">
      <GooseRuntimeProvider runtime={runtime}>
        <OIDCConsentPageView layout={layout} page={{ interaction: options.interaction ?? 'interaction-id' }} />
      </GooseRuntimeProvider>
    </GooseI18nProvider>,
  )
  return { decision, details, redirect, user: userEvent.setup() }
}

describe('OIDCConsentPageView', () => {
  it('loads trusted display details and translates known scopes', async () => {
    const { details } = renderConsent()

    expect(screen.getByRole('status').textContent).toContain('正在验证授权请求')
    expect(await screen.findByText('Goose Wiki')).toBeTruthy()
    expect(screen.getByText('确认你的身份')).toBeTruthy()
    expect(screen.getByText('读取名称、用户名和头像')).toBeTruthy()
    expect(screen.getByText('custom_scope')).toBeTruthy()
    expect(details).toHaveBeenCalledWith('interaction-id')
  })

  it('approves using only the trusted interaction and redirects to the result', async () => {
    const { decision, redirect, user } = renderConsent()
    await screen.findByText('Goose Wiki')

    await user.click(screen.getByRole('button', { name: '允许并继续' }))

    await waitFor(() => expect(decision).toHaveBeenCalledWith('interaction-id', 'approve'))
    expect(redirect).toHaveBeenCalledWith('https://client.example/callback')
  })

  it('shows an expired interaction without requesting details', async () => {
    const { details } = renderConsent({ interaction: '' })

    expect(await screen.findByText('授权请求已失效，请返回应用重新发起登录。')).toBeTruthy()
    expect(details).not.toHaveBeenCalled()
    expect(screen.getByRole('link', { name: '返回 GooseForum' }).getAttribute('href')).toBe('/')
  })

  it('shows details and decision failures without redirecting', async () => {
    renderConsent({ detailsError: new Error('授权请求已过期') })
    expect(await screen.findByText('授权请求已过期')).toBeTruthy()
    cleanup()

    const second = renderConsent()
    second.decision.mockRejectedValueOnce(new Error('无法处理授权请求'))
    await screen.findByText('Goose Wiki')
    await second.user.click(screen.getByRole('button', { name: '拒绝' }))

    expect(await screen.findByText('无法处理授权请求')).toBeTruthy()
    expect(second.redirect).not.toHaveBeenCalled()
    expect((screen.getByRole('button', { name: '拒绝' }) as HTMLButtonElement).disabled).toBe(false)
  })
})
